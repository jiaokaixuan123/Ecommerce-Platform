package service

// Service 层 / 业务逻辑
// 实现用户模块的核心业务逻辑，封装 Repository 层操作，处理业务规则

import (
	"context"
	"errors"
	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterReq：注册请求参数结构体
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

// LoginReq：登录请求参数结构体
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp：登录响应信息结构体
type LoginResp struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// UserService 接口：
type UserService interface {
	Register(ctx context.Context, req *RegisterReq) error
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
	GetUserInfo(ctx context.Context, userID uint) (*domain.User, error)
}

// userService 结构体：
type userService struct {
	repo      repository.UserRepository
	jwtSecret string	// jwt密钥
	jwtExpire int		// 过期时间
}

// NewUserService：创建 UserService 实例
func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpire int) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

// Register：接收 RegisterReq 的参数，封装注册逻辑
func (s *userService) Register(ctx context.Context, req *RegisterReq) error {
	// 检查用户名是否存在
	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrUserAlreadyExists))
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}

	return s.repo.Create(ctx, user)
}

// Login：登录
func (s *userService) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrUserNotFound))
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrPasswordWrong))
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID, "user", s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, err
	}

	return &LoginResp{Token: token, User: user}, nil
}

// GetUserInfo：获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, userID uint) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrUserNotFound))
		}
		return nil, err
	}
	return user, nil
}
