package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/ecommerce-platform/internal/seckill/service"
	"github.com/stretchr/testify/mock"
)

type MockSeckillService struct {
	mock.Mock
}

func (m *MockSeckillService) DoSeckill(ctx context.Context, req *service.DoSeckillReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockSeckillService) GetSeckillProduct(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillProduct), args.Error(1)
}

func (m *MockSeckillService) CreateSeckillProduct(ctx context.Context, req *service.CreateSeckillReq) (*domain.SeckillProduct, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillProduct), args.Error(1)
}

func (m *MockSeckillService) PrewarmStock(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
