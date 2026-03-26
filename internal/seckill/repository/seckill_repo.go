package repository

import (
	"context"

	"github.com/ecommerce-platform/internal/seckill/domain"
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
	// TODO: 实现查询逻辑
	panic("not implemented")
}

func (r *seckillRepository) Create(ctx context.Context, sp *domain.SeckillProduct) error {
	// TODO: 实现创建逻辑
	panic("not implemented")
}

func (r *seckillRepository) DecrStock(ctx context.Context, id uint, quantity int) error {
	// TODO: CAS 扣减 DB 库存，RowsAffected==0 时返回错误
	// UPDATE seckill_products SET remain_stock = remain_stock - ? WHERE id = ? AND remain_stock >= ?
	panic("not implemented")
}
