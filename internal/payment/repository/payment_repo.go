package repository

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/payment/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
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

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uint, from, to domain.PaymentStatus, thirdPartyNo string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Payment{}).
		Where("id = ? AND status = ?", id, from).
		Updates(map[string]interface{}{"status": to, "third_party_no": thirdPartyNo})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrPaymentDuplicate))
	}
	return nil
}
