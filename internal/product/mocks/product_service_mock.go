package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/product/domain"
	"github.com/ecommerce-platform/internal/product/service"
	"github.com/stretchr/testify/mock"
)

// MockProductService 是 ProductService 的 mock 实现
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, req *service.CreateProductReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockProductService) GetProduct(ctx context.Context, id uint) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) ListProducts(ctx context.Context, req *service.ListProductReq) (*service.ListProductResp, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ListProductResp), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id uint, req *service.UpdateProductReq) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) DeductStock(ctx context.Context, id uint, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}
