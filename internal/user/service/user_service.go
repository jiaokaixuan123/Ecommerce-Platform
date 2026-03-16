package service

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

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

type UserService interface {
	Register(ctx context.Context, req *RegisterReq) error
	Login(ctx context.Context, req *LoginReq) (*LoginResp, error)
	GetUserInfo(ctx context.Context, userID uint) (*domain.User, error)
}

type userService struct {
	repo      repository.UserRepository
	jwtSecret string
	jwtExpire int
}

func NewUserService(repo repository.UserRepository, jwtSecret string, jwtExpire int) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

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
