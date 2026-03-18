package service_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/mocks"
	"github.com/ecommerce-platform/internal/user/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	testJWTSecret = "test-secret"
	testJWTExpire = 24
)

func newService(repo *mocks.MockUserRepository) service.UserService {
	return service.NewUserService(repo, testJWTSecret, testJWTExpire)
}

// ---- Register ----

func TestRegister_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	// 模拟「用户注册时，先查询用户名是否存在（不存在）
	repo.On("GetByUsername", ctx, "alice").Return(nil, gorm.ErrRecordNotFound)
	// 验证要创建的用户用户名（验证密码不是原始明文）
	repo.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Username == "alice" && u.Password != "password123" // 密码已加密
	})).Return(nil)

	err := svc.Register(ctx, &service.RegisterReq{
		Username: "alice",
		Password: "password123",
		Email:    "alice@example.com",
	})

	assert.NoError(t, err)
	repo.AssertNumberOfCalls(t, "Create", 1)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByUsername", ctx, "alice").Return(&domain.User{Username: "alice"}, nil)

	err := svc.Register(ctx, &service.RegisterReq{
		Username: "alice",
		Password: "password123",
	})

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrUserAlreadyExists))
	repo.AssertNotCalled(t, "Create")
}

// ---- Login ----

func TestLogin_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	// 生成加密密码
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	// 构造预设用户
	user := &domain.User{ID: 1, Username: "alice", Password: string(hashed), Status: 1}

	repo.On("GetByUsername", ctx, "alice").Return(user, nil)

	resp, err := svc.Login(ctx, &service.LoginReq{
		Username: "alice",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)				// 验证 JWT 生成成功
	assert.Equal(t, user.ID, resp.User.ID)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByUsername", ctx, "ghost").Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Login(ctx, &service.LoginReq{Username: "ghost", Password: "pass"})

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrUserNotFound))
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &domain.User{ID: 1, Username: "alice", Password: string(hashed)}

	repo.On("GetByUsername", ctx, "alice").Return(user, nil)

	_, err := svc.Login(ctx, &service.LoginReq{Username: "alice", Password: "wrong"})

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrPasswordWrong))
}

// ---- GetUserInfo ----

func TestGetUserInfo_Success(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	user := &domain.User{ID: 1, Username: "alice"}
	repo.On("GetByID", ctx, uint(1)).Return(user, nil)

	result, err := svc.GetUserInfo(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, "alice", result.Username)
}

func TestGetUserInfo_NotFound(t *testing.T) {
	repo := new(mocks.MockUserRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.GetUserInfo(ctx, 99)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrUserNotFound))
}

