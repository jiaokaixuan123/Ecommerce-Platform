package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/ecommerce-platform/internal/seckill/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&domain.SeckillSaga{}, &domain.SeckillOutboxEvent{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

func TestCreateSagaWithOutbox_Success(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewSeckillSagaRepository(db)

	saga := &domain.SeckillSaga{
		SeckillID: 1,
		UserID:    2,
		Amount:    100,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}
	outbox := &domain.SeckillOutboxEvent{
		MaxRetry:  5,
		NextRunAt: time.Now(),
	}

	err := repo.CreateSagaWithOutbox(context.Background(), saga, outbox)
	assert.NoError(t, err)
	assert.NotZero(t, saga.ID)
	assert.NotZero(t, outbox.ID)
	assert.Equal(t, saga.ID, outbox.SagaID)
	assert.Equal(t, domain.OutboxStatusPending, outbox.Status)
}

func TestCreateSagaWithOutbox_Duplicate(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewSeckillSagaRepository(db)
	ctx := context.Background()

	firstSaga := &domain.SeckillSaga{
		SeckillID: 1,
		UserID:    2,
		Amount:    100,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}
	firstOutbox := &domain.SeckillOutboxEvent{NextRunAt: time.Now()}
	assert.NoError(t, repo.CreateSagaWithOutbox(ctx, firstSaga, firstOutbox))

	dupSaga := &domain.SeckillSaga{
		SeckillID: 1,
		UserID:    2,
		Amount:    100,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}
	dupOutbox := &domain.SeckillOutboxEvent{NextRunAt: time.Now()}
	err := repo.CreateSagaWithOutbox(ctx, dupSaga, dupOutbox)
	assert.Equal(t, pkgerrors.ErrSeckillRepeat, pkgerrors.CodeOf(err))
}

func TestOutboxClaimByID_OnlyOnce(t *testing.T) {
	db := newTestDB(t)
	outboxRepo := repository.NewSeckillOutboxRepository(db)
	ctx := context.Background()

	event := &domain.SeckillOutboxEvent{
		SagaID:     1,
		Topic:      "seckill.saga.execute",
		Status:     domain.OutboxStatusPending,
		RetryCount: 0,
		MaxRetry:   5,
		NextRunAt:  time.Now().Add(-time.Second),
	}
	assert.NoError(t, db.WithContext(ctx).Create(event).Error)

	got, claimed, err := outboxRepo.ClaimByID(ctx, event.ID)
	assert.NoError(t, err)
	assert.True(t, claimed)
	assert.NotNil(t, got)
	assert.Equal(t, domain.OutboxStatusProcessing, got.Status)

	got2, claimed2, err2 := outboxRepo.ClaimByID(ctx, event.ID)
	assert.NoError(t, err2)
	assert.False(t, claimed2)
	assert.Nil(t, got2)
}

func TestOutboxListDueIDs(t *testing.T) {
	db := newTestDB(t)
	outboxRepo := repository.NewSeckillOutboxRepository(db)
	ctx := context.Background()

	due := &domain.SeckillOutboxEvent{
		SagaID:    1,
		Topic:     "seckill.saga.execute",
		Status:    domain.OutboxStatusPending,
		NextRunAt: time.Now().Add(-time.Second),
		MaxRetry:  5,
	}
	future := &domain.SeckillOutboxEvent{
		SagaID:    2,
		Topic:     "seckill.saga.execute",
		Status:    domain.OutboxStatusPending,
		NextRunAt: time.Now().Add(time.Hour),
		MaxRetry:  5,
	}
	assert.NoError(t, db.WithContext(ctx).Create(due).Error)
	assert.NoError(t, db.WithContext(ctx).Create(future).Error)

	ids, err := outboxRepo.ListDueIDs(ctx, time.Now(), 10)
	assert.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, due.ID, ids[0])
}

func TestOutboxMarkStatusMethods(t *testing.T) {
	db := newTestDB(t)
	outboxRepo := repository.NewSeckillOutboxRepository(db)
	ctx := context.Background()

	event := &domain.SeckillOutboxEvent{
		SagaID:     1,
		Topic:      "seckill.saga.execute",
		Status:     domain.OutboxStatusPending,
		RetryCount: 0,
		MaxRetry:   5,
		NextRunAt:  time.Now(),
	}
	assert.NoError(t, db.WithContext(ctx).Create(event).Error)

	claimed, ok, err := outboxRepo.ClaimByID(ctx, event.ID)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NotNil(t, claimed)

	assert.NoError(t, outboxRepo.MarkRetry(ctx, event.ID, time.Now().Add(time.Minute), "e1"))
	var e1 domain.SeckillOutboxEvent
	assert.NoError(t, db.WithContext(ctx).Where("id = ?", event.ID).First(&e1).Error)
	assert.Equal(t, domain.OutboxStatusPending, e1.Status)
	assert.Equal(t, 1, e1.RetryCount)

	claimed2, ok2, err2 := outboxRepo.ClaimByID(ctx, event.ID)
	assert.NoError(t, err2)
	assert.True(t, ok2)
	assert.NotNil(t, claimed2)
	assert.NoError(t, outboxRepo.MarkDone(ctx, event.ID))

	var e2 domain.SeckillOutboxEvent
	assert.NoError(t, db.WithContext(ctx).Where("id = ?", event.ID).First(&e2).Error)
	assert.Equal(t, domain.OutboxStatusDone, e2.Status)

	assert.NoError(t, outboxRepo.MarkDead(ctx, event.ID, "dead"))
	var e3 domain.SeckillOutboxEvent
	assert.NoError(t, db.WithContext(ctx).Where("id = ?", event.ID).First(&e3).Error)
	assert.Equal(t, domain.OutboxStatusDead, e3.Status)
}
