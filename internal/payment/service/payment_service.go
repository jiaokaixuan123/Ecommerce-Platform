package service

import (
	"context"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/repository"
)

// ---- 请求/响应结构体 ----

// CreatePaymentReq 发起支付请求
type CreatePaymentReq struct {
	OrderID uint                  `json:"order_id" binding:"required"`
	UserID  uint                  `json:"user_id"`  // 从登录态获取
	Amount  int64                 `json:"amount" binding:"required,min=1"` // 支付金额（分）
	Channel domain.PaymentChannel `json:"channel" binding:"required"`
}

// PaymentCallbackReq 支付回调请求（第三方支付平台回调）
type PaymentCallbackReq struct {
	PaymentNo    string `json:"payment_no" binding:"required"`    // 平台支付流水号
	ThirdPartyNo string `json:"third_party_no" binding:"required"` // 第三方支付流水号
	Success      bool   `json:"success"`                          // 支付是否成功
}

// ---- 接口定义 ----

type PaymentService interface {
	// CreatePayment 发起支付，创建支付记录，返回支付参数（跳转 URL 或二维码等）
	// TODO Step 1: 校验订单是否存在且属于当前用户（需跨服务查询，MVP 阶段暂时跳过，直接信任前端传参）
	// TODO Step 2: 检查是否已有支付记录（幂等）
	// TODO Step 3: 生成平台支付流水号
	// TODO Step 4: 创建支付记录（状态：待支付）
	// TODO Step 5: 根据渠道调用对应支付网关（MVP 阶段用 mock）
	CreatePayment(ctx context.Context, req *CreatePaymentReq) (*domain.Payment, error)

	// HandleCallback 处理第三方支付回调
	// TODO Step 1: 根据 PaymentNo 查找支付记录
	// TODO Step 2: 幂等检查（已处理的回调直接返回）
	// TODO Step 3: 更新支付状态（CAS）
	// TODO Step 4: 通知订单服务更新订单状态（MVP 阶段直接调用，生产用消息队列）
	HandleCallback(ctx context.Context, req *PaymentCallbackReq) error

	// GetPaymentByOrderID 查询订单的支付记录
	GetPaymentByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error)
}

// ---- 实现 ----

type paymentService struct {
	paymentRepo repository.PaymentRepository
	// TODO: 生产环境需注入 OrderService 接口用于更新订单状态
	// orderService order.OrderService
}

func NewPaymentService(paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{paymentRepo: paymentRepo}
}

func (s *paymentService) CreatePayment(ctx context.Context, req *CreatePaymentReq) (*domain.Payment, error) {
	// TODO Step 1: 检查是否已有待支付/成功的支付记录（防重复）
	//   existing, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
	//   if err == nil { return existing, nil } // 已有记录，直接返回（幂等）

	// TODO Step 2: 生成平台流水号（格式：PAY + 时间戳 + UserID）
	//   paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), req.UserID)

	// TODO Step 3: 构造 Payment 实体并写库
	//   payment := &domain.Payment{ ... }
	//   if err := s.paymentRepo.Create(ctx, payment); err != nil { return nil, err }

	// TODO Step 4: 调用支付网关（mock 渠道直接返回成功）
	//   return payment, nil

	panic("not implemented")
}

func (s *paymentService) HandleCallback(ctx context.Context, req *PaymentCallbackReq) error {
	// TODO Step 1: 查找支付记录
	//   payment, err := s.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)

	// TODO Step 2: 幂等检查
	//   if payment.Status != domain.PaymentStatusPending { return nil }

	// TODO Step 3: 根据回调结果更新状态
	//   to := domain.PaymentStatusSuccess
	//   if !req.Success { to = domain.PaymentStatusFailed }
	//   if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, domain.PaymentStatusPending, to, req.ThirdPartyNo); err != nil { return err }

	// TODO Step 4: 通知订单服务（MVP 阶段可省略或直接调用 order service）

	panic("not implemented")
}

func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID uint) (*domain.Payment, error) {
	// TODO: 查询支付记录，找不到时返回 ErrNotFound
	panic("not implemented")
}
