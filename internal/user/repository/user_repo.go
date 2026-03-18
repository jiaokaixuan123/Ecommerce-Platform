package repository

// Repository 层 / 数据访问
// 封装数据库操作（基于 GORM），定义用户数据的 CRUD 接口，屏蔽底层数据库细节

import (
	"context"
	"github.com/ecommerce-platform/internal/user/domain"
	"gorm.io/gorm"
)

// UserRepository 接口：定义创建用户、按 ID / 用户名 / 邮箱 / 手机号查询用户、更新用户的方法
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// userRepository：GORM 的 db 实例
type userRepository struct {
	db *gorm.DB
}

// 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create：创建新用户
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID：通过 ID 索引用户
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return &user, err
}

// Update：更新用户数据
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
