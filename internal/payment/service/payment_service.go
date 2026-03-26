package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/repository"
)

// ---- 请求/响应结构体 ----

// CreatePaymentReq 发起支付请求
type CreatePaymentReq struct {
	OrderID uint                  `json:"order_id" binding:"required"`
	UserID  uint                  `json:"user_id"`                         // 从登录态获取
	Amount  int64                 `json:"amount" binding:"required,min=1"` // 支付金额（分）
	Channel domain.PaymentChannel `json:"channel" binding:"required"`
}

// PaymentCallbackReq 支付回调请求（第三方支付平台回调）
type PaymentCallbackReq struct {
	PaymentNo    string `json:"payment_no" binding:"required"`     // 平台支付流水号
	ThirdPartyNo string `json:"third_party_no" binding:"required"` // 第三方支付流水号
	Success      bool   `json:"success"`                           // 支付是否成功
}

// ---- 接口定义 ----

type PaymentService interface {
	// CreatePayment 创建订单
	CreatePayment(ctx context.Context, req *CreatePaymentReq) (*domain.Payment, error)

	// HandleCallback 处理订单回退
	HandleCallback(ctx context.Context, req *PaymentCallbackReq) error

	// GetPaymentByOrderID 查询订单的支付记录
	GetPaymentByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error)
}

// ---- 实现 ----

type paymentService struct {
	paymentRepo repository.PaymentRepository
	// 生产环境需注入 OrderService 接口用于更新订单状态
	// orderService order.OrderService
}

func NewPaymentService(paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{paymentRepo: paymentRepo}
}

func (s *paymentService) CreatePayment(ctx context.Context, req *CreatePaymentReq) (*domain.Payment, error) {
	// 幂等检验：若已存在支付记录，直接返回（防重复提交）
	existing, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err == nil && existing != nil {
		return existing, nil
	}
	// err != nil 时，仅 ErrRecordNotFound 允许继续创建，其他 DB 错误直接返回
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}

	// 生成平台流水号（格式：PAY + 时间戳 + UserID）
	paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), req.UserID)

	// 构造 Payment 实体并写库
	payment := &domain.Payment{
		PaymentNo:    paymentNo,
		OrderID:      req.OrderID,
		UserID:       req.UserID,
		Amount:       req.Amount,
		Channel:      req.Channel,
		Status:       domain.PaymentStatusPending,
		ThirdPartyNo: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// 调用支付网关（mock 渠道直接返回成功）
	return payment, nil
}

func (s *paymentService) HandleCallback(ctx context.Context, req *PaymentCallbackReq) error {
	// 查找支付记录
	payment, err := s.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		return err
	}

	// 幂等检查
	if payment.Status != domain.PaymentStatusPending {
		return nil
	}

	// 根据回调结果更新状态
	to := domain.PaymentStatusSuccess
	if !req.Success {
		to = domain.PaymentStatusFailed
	}
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentStatusPending, to, req.ThirdPartyNo); err != nil {
		return err
	}

	// 通知订单服务（MVP 阶段可省略或直接调用 order service）
	return nil
}

func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error) {
	payment, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return payment, nil
}
