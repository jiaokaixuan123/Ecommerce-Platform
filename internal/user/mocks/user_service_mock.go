package mocks

import (
	"context"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/service"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *service.RegisterReq) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserService) Login(ctx context.Context, req *service.LoginReq) (*service.LoginResp, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResp), args.Error(1)
}

func (m *MockUserService) GetUserInfo(ctx context.Context, userID uint) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
