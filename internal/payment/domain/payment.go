package domain

import "time"

// PaymentStatus 支付状态
type PaymentStatus int8

const (
	PaymentStatusPending  PaymentStatus = 1 // 待支付
	PaymentStatusSuccess  PaymentStatus = 2 // 支付成功
	PaymentStatusFailed   PaymentStatus = 3 // 支付失败
	PaymentStatusRefunded PaymentStatus = 4 // 已退款
)

// PaymentChannel 支付渠道
type PaymentChannel string

const (
	PaymentChannelAlipay PaymentChannel = "alipay" // 支付宝
	PaymentChannelWechat PaymentChannel = "wechat" // 微信支付
	PaymentChannelMock   PaymentChannel = "mock"   // 模拟支付（开发/测试用）
)

// Payment 支付记录
type Payment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PaymentNo     string         `gorm:"uniqueIndex;size:32;not null" json:"payment_no"` // 平台支付流水号
	OrderID       uint           `gorm:"not null;index" json:"order_id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	Amount        int64          `gorm:"not null" json:"amount"`                        // 支付金额（分）
	Channel       PaymentChannel `gorm:"size:20;not null" json:"channel"`
	Status        PaymentStatus  `gorm:"not null;default:1" json:"status"`
	ThirdPartyNo  string         `gorm:"size:64" json:"third_party_no"`  // 第三方支付流水号（回调后填充）
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }
