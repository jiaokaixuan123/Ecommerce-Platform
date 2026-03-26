package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/stretchr/testify/mock"
)

type MockSeckillRepository struct {
	mock.Mock
}

func (m *MockSeckillRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillProduct), args.Error(1)
}

func (m *MockSeckillRepository) Create(ctx context.Context, sp *domain.SeckillProduct) error {
	args := m.Called(ctx, sp)
	return args.Error(0)
}

func (m *MockSeckillRepository) DecrStock(ctx context.Context, id uint, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}
