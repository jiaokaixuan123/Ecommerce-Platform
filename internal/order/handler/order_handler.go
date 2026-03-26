package handler

import (
	"strconv"

	"github.com/ecommerce-platform/internal/order/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: svc}
}

// CreateOrder POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	req.UserID = middleware.GetUserID(c)

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}
	response.Success(c, order)
}

// GetOrder GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderId, _ :=strconv.ParseUint(c.Param("id"), 10, 64)
	userId := middleware.GetUserID(c)
	order, err := h.orderService.GetOrderDetail(c.Request.Context(), userId, uint(orderId))
	if err != nil {
		response.Fail(c, pkgerrors.ErrOrderNotFound, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	userId := middleware.GetUserID(c)
	pageint, _ := strconv.Atoi(page)
	pageSizeint, _ := strconv.Atoi(pageSize)
	resp, err := h.orderService.ListUserOrders(c.Request.Context(), userId, pageint, pageSizeint)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	response.Success(c, resp)
}

// CancelOrder POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderId, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userId := middleware.GetUserID(c)
	err := h.orderService.CancelOrder(c.Request.Context(), userId, uint(orderId))
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	response.Success(c, nil)
}
