package handler

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册秒杀模块路由
// secret: JWT 签名密钥
func RegisterRoutes(r *gin.RouterGroup, h *SeckillHandler, secret string) {
	seckill := r.Group("/seckill")

	// 需要登录
	auth := seckill.Group("", middleware.JWT(secret))
	{
		auth.POST("/:id", h.DoSeckill)
		auth.GET("/:id", h.GetSeckillProduct)
	}

	// 管理员接口（生产环境应加 role 校验中间件）
	admin := seckill.Group("/admin", middleware.JWT(secret))
	{
		admin.POST("", h.CreateSeckillProduct)
		admin.POST("/:id/prewarm", h.PrewarmStock)
	}
}
