package repository

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/order/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// OrderRepository 订单主表数据访问接口
type OrderRepository interface {
	// Create 创建订单（同时写入 order + order_items，需在事务中执行）
	Create(ctx context.Context, order *domain.Order, items []*domain.OrderItem) error

	// GetByID 根据订单 ID 查询
	GetByID(ctx context.Context, id uint) (*domain.Order, error)

	// GetByOrderNo 根据业务订单号查询
	GetByOrderNo(ctx context.Context, orderNo string) (*domain.Order, error)

	// ListByUserID 分页查询用户的订单列表
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*domain.Order, int64, error)

	// UpdateStatus 更新订单状态（CAS：只在当前状态匹配时才更新，防止并发问题）
	UpdateStatus(ctx context.Context, id uint, from, to domain.OrderStatus) error
}

// OrderItemRepository 订单商品项数据访问接口
type OrderItemRepository interface {
	// ListByOrderID 查询订单下的所有商品项
	ListByOrderID(ctx context.Context, orderID uint) ([]*domain.OrderItem, error)
}

type orderRepository struct {
	db *gorm.DB
}

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}


func (r *orderRepository) Create(ctx context.Context, order *domain.Order, items []*domain.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(order).Error; err != nil {
        return err
    }
    for _, item := range items {
        item.OrderID = order.ID
    }
    return tx.Create(&items).Error
})
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*domain.Order, int64, error) {
	var (
		orders []*domain.Order
		total  int64
	)

	query := r.db.WithContext(ctx).Model(&domain.Order{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil

}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, from, to domain.OrderStatus) error {
    result := r.db.WithContext(ctx).
        Model(&domain.Order{}).
        Where("id = ? AND status = ?", id, from).
        Update("status", to)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        // 没有行被更新：说明订单不存在，或当前状态已经不是 from（并发场景）
        return errors.New(pkgerrors.Msg(pkgerrors.ErrOrderStatusInvalid))
    }
    return nil
}

func (r *orderItemRepository) ListByOrderID(ctx context.Context, orderID uint) ([]*domain.OrderItem, error) {
	var orderItems []*domain.OrderItem
	query := r.db.WithContext(ctx).Model(&domain.OrderItem{}).Where("order_id = ?", orderID)
	if err := query.Find(&orderItems).Error; err != nil {
		return nil, err
	}

	return orderItems, nil
}
