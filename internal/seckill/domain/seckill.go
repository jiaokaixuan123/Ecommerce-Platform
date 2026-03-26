package domain

import "time"

// SeckillStatus 秒杀活动状态
type SeckillStatus int8

const (
	SeckillStatusPending  SeckillStatus = 1 // 未开始
	SeckillStatusActive   SeckillStatus = 2 // 进行中
	SeckillStatusFinished SeckillStatus = 3 // 已结束
	SeckillStatusDisabled SeckillStatus = 4 // 已下架
)

// SeckillProduct 秒杀商品活动表
type SeckillProduct struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	ProductID    uint          `gorm:"not null;index" json:"product_id"`
	ProductName  string        `gorm:"size:100;not null" json:"product_name"` // 快照
	ProductImage string        `gorm:"size:255" json:"product_image"`         // 快照
	SeckillPrice int64         `gorm:"not null" json:"seckill_price"`         // 秒杀价（分）
	TotalStock   int           `gorm:"not null" json:"total_stock"`           // 总库存
	RemainStock  int           `gorm:"not null" json:"remain_stock"`          // DB 剩余库存（异步落库后更新）
	StartAt      time.Time     `gorm:"not null" json:"start_at"`
	EndAt        time.Time     `gorm:"not null" json:"end_at"`
	Status       SeckillStatus `gorm:"not null;default:1" json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

func (SeckillProduct) TableName() string { return "seckill_products" }

// SeckillOrder 秒杀排队记录（异步落库前的中间状态）
type SeckillOrder struct {
	SeckillID uint  `json:"seckill_id"`
	UserID    uint  `json:"user_id"`
	Amount    int64 `json:"amount"` // 秒杀价
}
