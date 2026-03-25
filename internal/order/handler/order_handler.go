package handler

import (
	"strconv"

	"github.com/ecommerce-platform/internal/order/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: svc}
}

// CreateOrder POST /api/v1/orders
// TODO Step 1：
//   - ShouldBindJSON 解析请求体到 service.CreateOrderReq
//   - 从 middleware.GetUserID(c) 获取 userID 写入 req.UserID
//   - 调用 h.orderService.CreateOrder
//   - 成功返回创建的订单对象
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// TODO
}

// GetOrder GET /api/v1/orders/:id
// TODO Step 2：
//   - 用 strconv.ParseUint 解析路径参数 id
//   - 调用 h.orderService.GetOrderDetail(ctx, userID, orderID)
//   - 注意越权时返回 ErrForbidden
func (h *OrderHandler) GetOrder(c *gin.Context) {
	// TODO
	_ = strconv.ParseUint
}

// ListOrders GET /api/v1/orders?page=1&page_size=10
// TODO Step 3：
//   - 用 c.DefaultQuery 解析分页参数，page 默认 1，page_size 默认 10
//   - 调用 h.orderService.ListUserOrders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// TODO
}

// CancelOrder POST /api/v1/orders/:id/cancel
// TODO Step 4：
//   - 解析路径参数 id
//   - 调用 h.orderService.CancelOrder(ctx, userID, orderID)
//   - 状态不合法时返回 ErrOrderStatusInvalid
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	// TODO
	_ = pkgerrors.ErrOrderStatusInvalid
}

func RegisterRoutes(r *gin.Engine, h *OrderHandler, jwtSecret string) {
	auth := middleware.JWT(jwtSecret)

	api := r.Group("/api/v1")
	orders := api.Group("/orders", auth)
	{
		orders.POST("", h.CreateOrder)             // 创建订单
		orders.GET("", h.ListOrders)               // 订单列表
		orders.GET("/:id", h.GetOrder)             // 订单详情
		orders.POST("/:id/cancel", h.CancelOrder)  // 取消订单
	}
}
