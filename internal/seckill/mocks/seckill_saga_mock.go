package mocks

import (
	"context"
	"time"

	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/stretchr/testify/mock"
)

type MockSeckillSagaRepository struct {
	mock.Mock
}

func (m *MockSeckillSagaRepository) CreateSagaWithOutbox(ctx context.Context, saga *domain.SeckillSaga, outbox *domain.SeckillOutboxEvent) error {
	args := m.Called(ctx, saga, outbox)
	return args.Error(0)
}

func (m *MockSeckillSagaRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillSaga, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillSaga), args.Error(1)
}

func (m *MockSeckillSagaRepository) Save(ctx context.Context, saga *domain.SeckillSaga) error {
	args := m.Called(ctx, saga)
	return args.Error(0)
}

type MockSeckillOutboxRepository struct {
	mock.Mock
}

func (m *MockSeckillOutboxRepository) ListDueIDs(ctx context.Context, now time.Time, limit int) ([]uint, error) {
	args := m.Called(ctx, now, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint), args.Error(1)
}

func (m *MockSeckillOutboxRepository) ClaimByID(ctx context.Context, id uint) (*domain.SeckillOutboxEvent, bool, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*domain.SeckillOutboxEvent), args.Bool(1), args.Error(2)
}

func (m *MockSeckillOutboxRepository) MarkDone(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSeckillOutboxRepository) MarkRetry(ctx context.Context, id uint, nextRunAt time.Time, lastErr string) error {
	args := m.Called(ctx, id, nextRunAt, lastErr)
	return args.Error(0)
}

func (m *MockSeckillOutboxRepository) MarkDead(ctx context.Context, id uint, lastErr string) error {
	args := m.Called(ctx, id, lastErr)
	return args.Error(0)
}
