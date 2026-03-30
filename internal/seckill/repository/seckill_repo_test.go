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

func newSeckillRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:seckill_repo_" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&domain.SeckillProduct{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

func TestSeckillRepository_GetByID_NotFound(t *testing.T) {
	db := newSeckillRepoDB(t)
	repo := repository.NewSeckillRepository(db)
	_, err := repo.GetByID(context.Background(), 1)
	assert.Equal(t, pkgerrors.ErrProductNotFound, pkgerrors.CodeOf(err))
}

func TestSeckillRepository_CreateAndGet(t *testing.T) {
	db := newSeckillRepoDB(t)
	repo := repository.NewSeckillRepository(db)
	ctx := context.Background()
	p := &domain.SeckillProduct{
		ProductID:    1,
		ProductName:  "p",
		SeckillPrice: 100,
		TotalStock:   10,
		RemainStock:  10,
		StartAt:      time.Now(),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusPending,
	}
	assert.NoError(t, repo.Create(ctx, p))
	got, err := repo.GetByID(ctx, p.ID)
	assert.NoError(t, err)
	assert.Equal(t, p.ProductID, got.ProductID)
}

func TestSeckillRepository_DecrStock(t *testing.T) {
	db := newSeckillRepoDB(t)
	repo := repository.NewSeckillRepository(db)
	ctx := context.Background()
	p := &domain.SeckillProduct{
		ProductID:    1,
		ProductName:  "p",
		SeckillPrice: 100,
		TotalStock:   1,
		RemainStock:  1,
		StartAt:      time.Now(),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusPending,
	}
	assert.NoError(t, repo.Create(ctx, p))
	assert.NoError(t, repo.DecrStock(ctx, p.ID, 1))
	err := repo.DecrStock(ctx, p.ID, 1)
	assert.Equal(t, pkgerrors.ErrProductOutOfStock, pkgerrors.CodeOf(err))
}
