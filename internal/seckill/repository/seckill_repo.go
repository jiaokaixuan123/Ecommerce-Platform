package repository

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/seckill/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// SeckillRepository 秒杀商品数据访问接口
type SeckillRepository interface {
	// GetByID 查询秒杀商品详情
	GetByID(ctx context.Context, id uint) (*domain.SeckillProduct, error)

	// Create 创建秒杀活动（管理员）
	Create(ctx context.Context, sp *domain.SeckillProduct) error

	// DecrStock 同步扣减 DB 库存（CAS，异步 worker 调用）
	DecrStock(ctx context.Context, id uint, quantity int) error
}

type seckillRepository struct {
	db *gorm.DB
}

func NewSeckillRepository(db *gorm.DB) SeckillRepository {
	return &seckillRepository{db: db}
}

func (r *seckillRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	var seckillProduct domain.SeckillProduct
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&seckillProduct).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.New(pkgerrors.ErrProductNotFound)
		}
		return nil, err
	}
	return &seckillProduct, nil
}

func (r *seckillRepository) Create(ctx context.Context, sp *domain.SeckillProduct) error {
	return r.db.WithContext(ctx).Create(sp).Error
}

func (r *seckillRepository) DecrStock(ctx context.Context, id uint, quantity int) error {
	// CAS 核心：WHERE 条件保证原子性，不会超卖
	result := r.db.WithContext(ctx).
		Model(&domain.SeckillProduct{}).
		Where("id = ? AND remain_stock >= ?", id, quantity).
		UpdateColumn("remain_stock", gorm.Expr("remain_stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	// 没有行受影响 → 库存不足/商品不存在 → 扣减失败
	if result.RowsAffected == 0 {
		return pkgerrors.New(pkgerrors.ErrProductOutOfStock)
	}

	return nil
}
