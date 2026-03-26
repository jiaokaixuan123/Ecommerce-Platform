package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/stretchr/testify/mock"
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	args := m.Called(ctx, paymentNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, id uint, from, to domain.PaymentStatus, thirdPartyNo string) error {
	args := m.Called(ctx, id, from, to, thirdPartyNo)
	return args.Error(0)
}
