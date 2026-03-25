package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *CartHandler, jwtSecret string) {
	auth := middleware.JWT(jwtSecret)

	api := r.Group("/api/v1")
	cart := api.Group("/cart", auth) // 购物车所有接口均需登录
	{
		cart.GET("", h.GetUserCart)                        // 获取购物车
		cart.POST("/items", h.AddCartItem)                  // 加入购物车
		cart.PUT("/items/:id/quantity", h.UpdateCartItemQuantity) // 修改数量
		cart.PUT("/items/:id/selected", h.UpdateCartItemSelected) // 修改选中
		cart.DELETE("/items", h.DeleteCartItem)             // 删除（按 item_id 或 product_id）
		cart.DELETE("", h.ClearCart)                        // 清空购物车
	}
}
