package domain

import "time"

// OrderStatus 订单状态
type OrderStatus int8

const (
	OrderStatusPending   OrderStatus = 1 // 待支付
	OrderStatusPaid      OrderStatus = 2 // 已支付
	OrderStatusShipping  OrderStatus = 3 // 配送中
	OrderStatusCompleted OrderStatus = 4 // 已完成
	OrderStatusCancelled OrderStatus = 5 // 已取消
)

// Order 订单主表
type Order struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	OrderNo       string      `gorm:"uniqueIndex;size:32;not null" json:"order_no"`  // 业务订单号（如 ORD20240101XXXXXX）
	UserID        uint        `gorm:"not null;index" json:"user_id"`
	TotalAmount   int64       `gorm:"not null" json:"total_amount"`   // 订单总金额（分）
	Status        OrderStatus `gorm:"not null;default:1" json:"status"`
	Remark        string      `gorm:"size:255" json:"remark"`         // 用户备注
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

// OrderItem 订单商品项（快照）
type OrderItem struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	OrderID      uint   `gorm:"not null;index" json:"order_id"`
	ProductID    uint   `gorm:"not null" json:"product_id"`
	MerchantID   uint   `gorm:"not null" json:"merchant_id"`
	ProductName  string `gorm:"size:100;not null" json:"product_name"`  // 下单时快照
	ProductImage string `gorm:"size:255" json:"product_image"`
	Price        int64  `gorm:"not null" json:"price"`    // 下单时单价（分，快照）
	Quantity     int    `gorm:"not null" json:"quantity"` // 购买数量
	Subtotal     int64  `gorm:"not null" json:"subtotal"` // 小计 = price * quantity
}

func (OrderItem) TableName() string { return "order_items" }
