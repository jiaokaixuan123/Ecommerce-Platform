package domain

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type CartItem struct {
	// 基础元信息
	ID        uint           `gorm:"primarykey" json:"id"` // 购物车项ID
	CreatedAt time.Time      `json:"created_at"`           // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`       // 软删除标识

	// 核心必备字段
	CartID     uint `gorm:"index;not null" json:"cart_id"`         // 关联购物车主ID
	ProductID  uint `gorm:"index;not null" json:"product_id"`      // 关联商品ID
	MerchantID uint `gorm:"index;not null" json:"merchant_id"`     // 关联商家ID
	Quantity   int  `gorm:"not null;default:1" json:"quantity"`    // 商品数量（≥1）
	Selected   bool `gorm:"not null;default:true" json:"selected"` // 是否选中（结算时是否包含）

	// 商品快照字段（避免商品信息变更导致购物车数据不一致）
	ProductName  string `gorm:"not null" json:"product_name"`  // 加购时的商品名称
	ProductPrice int64  `gorm:"not null" json:"product_price"` // 加购时的商品单价
	ProductImage string `json:"product_image"`                 // 加购时的商品主图
}
