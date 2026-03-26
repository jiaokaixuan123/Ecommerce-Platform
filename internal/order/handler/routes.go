package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

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