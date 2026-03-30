package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/order/domain"
	ordersvc "github.com/ecommerce-platform/internal/order/service"
	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, req *ordersvc.CreateOrderReq) (*domain.Order, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderDetail(ctx context.Context, userID, orderID uint) (*ordersvc.OrderDetailResp, error) {
	args := m.Called(ctx, userID, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordersvc.OrderDetailResp), args.Error(1)
}

func (m *MockOrderService) ListUserOrders(ctx context.Context, userID uint, page, pageSize int) (*ordersvc.ListOrderResp, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordersvc.ListOrderResp), args.Error(1)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, userID, orderID uint) error {
	args := m.Called(ctx, userID, orderID)
	return args.Error(0)
}

func (m *MockOrderService) ConfirmOrder(ctx context.Context, orderID uint) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}
