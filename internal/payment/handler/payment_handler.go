package handler

import (
	"github.com/ecommerce-platform/internal/payment/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: svc}
}

// CreatePayment POST /api/v1/payments
// 用户发起支付
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req service.CreatePaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	req.UserID = middleware.GetUserID(c)

	// TODO: 调用 h.paymentService.CreatePayment，成功返回支付记录
	panic("not implemented")
}

// HandleCallback POST /api/v1/payments/callback
// 第三方支付平台异步回调，不需要登录态
func (h *PaymentHandler) HandleCallback(c *gin.Context) {
	var req service.PaymentCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	// TODO: 调用 h.paymentService.HandleCallback，成功返回 Success
	panic("not implemented")
}

// GetPaymentByOrderID GET /api/v1/payments/order/:order_id
// 查询订单支付状态
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	// TODO: 解析 order_id 路径参数，调用 GetPaymentByOrderID，返回支付记录
	panic("not implemented")
}


