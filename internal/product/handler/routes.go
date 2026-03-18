package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册商品服务路由
func RegisterRoutes(r *gin.Engine, h *ProductHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	auth := middleware.JWT(jwtSecret)
	merchantOnly := middleware.RoleAuth("merchant", "admin")

	products := api.Group("/products")
	{
		products.GET("", h.ListProducts)           // 公开：商品列表
		products.GET("/:id", h.GetProduct)         // 公开：商品详情

		products.POST("", auth, merchantOnly, h.CreateProduct)       // 仅商家/管理员
		products.PUT("/:id", auth, merchantOnly, h.UpdateProduct)    // 仅商家/管理员
		products.DELETE("/:id", auth, merchantOnly, h.DeleteProduct) // 仅商家/管理员
	}
}
