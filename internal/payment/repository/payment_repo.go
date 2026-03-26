package repository

import (
	"context"

	"github.com/ecommerce-platform/internal/payment/domain"
)

// PaymentRepository 支付记录数据访问接口
type PaymentRepository interface {
	// Create 创建支付记录
	Create(ctx context.Context, payment *domain.Payment) error

	// GetByPaymentNo 根据平台支付流水号查询
	GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error)

	// GetByOrderID 根据订单 ID 查询支付记录
	GetByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error)

	// UpdateStatus 更新支付状态（CAS：只在当前状态匹配时才更新）
	UpdateStatus(ctx context.Context, id uint, from, to domain.PaymentStatus, thirdPartyNo string) error
}
