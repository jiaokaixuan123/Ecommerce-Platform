package repository

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/cart/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

type CartRepository interface {
	// GetOrCreateByUserID 获取用户购物车（无则创建）
	GetOrCreateByUserID(ctx context.Context, userID uint) (*domain.Cart, error)
}

type CartItemRepository interface {
	// Add 新增购物车项（已存在则更新数量）
	Add(ctx context.Context, item *domain.CartItem) error

	// ListByCartID 根据购物车ID查询所有商品项
	ListByCartID(ctx context.Context, cartID uint) ([]*domain.CartItem, error)

	// GetItemByID 确认 item 存在且属于当前用户
	GetItemByID(ctx context.Context, itemID uint) (*domain.CartItem, error)

	// UpdateQuantity 更新购物车项数量（原子操作）
	UpdateQuantity(ctx context.Context, itemID uint, quantity int) error

	// UpdateSelected 更新购物车项选中状态
	UpdateSelected(ctx context.Context, itemID uint, selected bool) error

	// Delete 删除购物车项（软删除）
	Delete(ctx context.Context, itemID uint) error

	// DeleteByCartIDAndProductID 根据购物车ID+商品ID删除项
	DeleteByCartIDAndProductID(ctx context.Context, cartID, productID uint) error

	// ClearByCartID 下单成功后清空购物车
	ClearByCartID(ctx context.Context, cartID uint) error
}

// ---------------------------------------

type cartRepository struct {
	db *gorm.DB
}

type cartItemRepository struct {
	db *gorm.DB
}

// ---------------------------------------

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func NewCartItemRepository(db *gorm.DB) CartItemRepository {
	return &cartItemRepository{db: db}
}

// ---------------------------------------
func (r *cartRepository) GetOrCreateByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	// 先查询
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error
	if err == nil {
		// 已存在，直接返回
		return &cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 不存在，创建
	cart = domain.Cart{UserID: userID}
	if err := r.db.WithContext(ctx).Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartItemRepository) Add(ctx context.Context, item *domain.CartItem) error {
	var existing domain.CartItem
	err := r.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", item.CartID, item.ProductID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在，直接创建
		return r.db.WithContext(ctx).Create(item).Error
	}
	if err != nil {
		return err
	}

	// 已存在，累加数量
	return r.db.WithContext(ctx).Model(&existing).
		Update("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error
}

func (r *cartItemRepository) ListByCartID(ctx context.Context, cartID uint) ([]*domain.CartItem, error) {
	var items []*domain.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ? ", cartID).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartItemRepository) GetItemByID(ctx context.Context, itemID uint) (*domain.CartItem, error) {
	var item *domain.CartItem
	err := r.db.WithContext(ctx).First(&item, itemID).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *cartItemRepository) UpdateQuantity(ctx context.Context, itemID uint, quantity int) error {
	var item *domain.CartItem
	err := r.db.WithContext(ctx).Where("item_id = ? ", itemID).First(&item).Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Where("item_id = ? ", itemID).
		Update("quantity", quantity).Error
}

func (r *cartItemRepository) UpdateSelected(ctx context.Context, itemID uint, selected bool) error {
	return r.db.WithContext(ctx).Where("item_id = ? ", itemID).
		Update("selected", selected).Error
}

func (r *cartItemRepository) Delete(ctx context.Context, itemID uint) error {
	return r.db.WithContext(ctx).Delete(&domain.CartItem{}, itemID).Error
}

func (r *cartItemRepository) DeleteByCartIDAndProductID(ctx context.Context, cartID, productID uint) error {
	result := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ? ", cartID, productID).Delete(&domain.CartItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrNotFound))
	}
	return nil
}

func (r *cartItemRepository) ClearByCartID(ctx context.Context, cartID uint) error {
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&domain.CartItem{}).Error
}
