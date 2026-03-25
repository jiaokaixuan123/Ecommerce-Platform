package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/cart/domain"
	"github.com/stretchr/testify/mock"
)

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) GetOrCreateByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cart), args.Error(1)
}

type MockCartItemRepository struct {
	mock.Mock
}

func (m *MockCartItemRepository) Add(ctx context.Context, item *domain.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartItemRepository) ListByCartID(ctx context.Context, cartID uint) ([]*domain.CartItem, error) {
	args := m.Called(ctx, cartID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CartItem), args.Error(1)
}

func (m *MockCartItemRepository) GetItemByID(ctx context.Context, itemID uint) (*domain.CartItem, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CartItem), args.Error(1)
}

func (m *MockCartItemRepository) UpdateQuantity(ctx context.Context, itemID uint, quantity int) error {
	args := m.Called(ctx, itemID, quantity)
	return args.Error(0)
}

func (m *MockCartItemRepository) UpdateSelected(ctx context.Context, itemID uint, selected bool) error {
	args := m.Called(ctx, itemID, selected)
	return args.Error(0)
}

func (m *MockCartItemRepository) Delete(ctx context.Context, itemID uint) error {
	args := m.Called(ctx, itemID)
	return args.Error(0)
}

func (m *MockCartItemRepository) DeleteByCartIDAndProductID(ctx context.Context, cartID, productID uint) error {
	args := m.Called(ctx, cartID, productID)
	return args.Error(0)
}

func (m *MockCartItemRepository) ClearByCartID(ctx context.Context, cartID uint) error {
	args := m.Called(ctx, cartID)
	return args.Error(0)
}
