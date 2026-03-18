package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

// mock对象
type MockUserRepository struct {
	mock.Mock
}

// 
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	// Called() 记录方法调用，并返回预设的返回值
	args := m.Called(ctx, user)
	// 返回错误值
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	// 处理返回值：如果第一个返回值为 nil，返回 nil + 第二个返回值（error）
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// 否则返回第一个返回值（*domain.User） + 第二个返回值（error）
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
