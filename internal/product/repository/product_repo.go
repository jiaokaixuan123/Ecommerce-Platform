package repository

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/product/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// ProductRepository 定义商品数据访问接口
type ProductRepository interface {
	// 创建商品
	Create(ctx context.Context, product *domain.Product) error

	// 根据 ID 查询商品
	GetByID(ctx context.Context, id uint) (*domain.Product, error)

	// 分页查询商品列表（支持按分类过滤）
	// offset: 跳过的记录数, limit: 每页数量, categoryID: 0 表示不过滤
	List(ctx context.Context, offset, limit int, categoryID uint) ([]*domain.Product, int64, error)

	// 更新商品信息
	Update(ctx context.Context, product *domain.Product) error

	// 删除商品（软删除或直接删除均可）
	Delete(ctx context.Context, id uint) error

	// 扣减库存（需要保证原子性，防止超卖）
	// 提示：使用 UPDATE products SET stock = stock - quantity WHERE id = ? AND stock >= quantity
	DeductStock(ctx context.Context, id uint, quantity int) error
}

// productRepository 是 ProductRepository 的具体实现
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建 productRepository 实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// 实现 Create
func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// 实现 GetByID
func (r *productRepository) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	// First 方法：查询不到会返回 gorm.ErrRecordNotFound
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// 实现 List（分页 + 可选分类过滤）
func (r *productRepository) List(ctx context.Context, offset, limit int, categoryID uint) ([]*domain.Product, int64, error) {
	var (
		products []*domain.Product
		total    int64
	)
	// 构建查询条件
	query := r.db.WithContext(ctx).Model(&domain.Product{})
	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询数据
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// 实现 Update
func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// 实现 Delete
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Product{}, id).Error
}

// 实现 DeductStock（原子扣减，防止超卖）
func (r *productRepository) DeductStock(ctx context.Context, id uint, quantity int) error {
	result := r.db.WithContext(ctx).Model(&domain.Product{}).
		Where("id = ? AND stock >= ?", id, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		var product domain.Product
		err := r.db.WithContext(ctx).First(&product, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 商品不存在
			return errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound))
		} else {
			// 库存不足
			return errors.New(pkgerrors.Msg(pkgerrors.ErrProductOutOfStock))
		}
	}

	return nil
}
