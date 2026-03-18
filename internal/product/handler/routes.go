package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册商品服务路由
func RegisterRoutes(r *gin.Engine, h *ProductHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	auth := middleware.JWT(jwtSecret)

	products := api.Group("/products")
	{
		products.GET("", h.ListProducts)            // 公开：商品列表
		products.GET("/:id", h.GetProduct)          // 公开：商品详情

		// 以下需要登录（TODO: 生产环境还需要管理员角色校验）
		products.POST("", auth, h.CreateProduct)        // 创建商品
		products.PUT("/:id", auth, h.UpdateProduct)     // 更新商品
		products.DELETE("/:id", auth, h.DeleteProduct)  // 删除商品
	}
}
