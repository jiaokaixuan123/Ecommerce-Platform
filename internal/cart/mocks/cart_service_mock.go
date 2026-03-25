package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/cart/service"
	"github.com/stretchr/testify/mock"
)

type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) AddCartItem(ctx context.Context, req *service.AddCartItemReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockCartService) GetUserCart(ctx context.Context, userID uint) (*service.CartVO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CartVO), args.Error(1)
}

func (m *MockCartService) UpdateCartItemQuantity(ctx context.Context, req *service.UpdateCartItemQuantityReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockCartService) UpdateCartItemSelected(ctx context.Context, req *service.UpdateCartItemSelectedReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockCartService) DeleteCartItem(ctx context.Context, req *service.DeleteCartItemReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockCartService) ClearCart(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
