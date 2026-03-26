package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)
// RegisterRoutes 注册支付服务路由
func RegisterRoutes(r *gin.Engine, h *PaymentHandler, jwtSecret string) {
	auth := middleware.JWT(jwtSecret)

	api := r.Group("/api/v1/payments")
	{
		api.POST("", auth, h.CreatePayment)                          // 发起支付（需登录）
		api.POST("/callback", h.HandleCallback)                      // 第三方回调（无需登录）
		api.GET("/order/:order_id", auth, h.GetPaymentByOrderID)     // 查询支付状态（需登录）
	}
}