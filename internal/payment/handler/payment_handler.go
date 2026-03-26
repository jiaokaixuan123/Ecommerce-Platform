package handler

import (
	"strconv"

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

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, payment)
}

// HandleCallback POST /api/v1/payments/callback
// 第三方支付平台异步回调，不需要登录态
func (h *PaymentHandler) HandleCallback(c *gin.Context) {
	var req service.PaymentCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	err := h.paymentService.HandleCallback(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetPaymentByOrderID GET /api/v1/payments/order/:order_id
// 查询订单支付状态
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderId, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	payment, err := h.paymentService.GetPaymentByOrderID(c.Request.Context(), uint(orderId))
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, payment)
}


