package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/service"
	"github.com/stretchr/testify/mock"
)

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, req *service.CreatePaymentReq) (*domain.Payment, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentService) HandleCallback(ctx context.Context, req *service.PaymentCallbackReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockPaymentService) GetPaymentByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}
