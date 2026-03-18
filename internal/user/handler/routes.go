package handler

// 路由注册
// 定义用户模块的 API 路由，绑定 Handler 层的处理函数，并配置认证中间件

import (
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes : 注册路由
func RegisterRoutes(r *gin.Engine, h *UserHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	{
		user := api.Group("/user")
		{
			user.POST("/register", h.Register)
			user.POST("/login", h.Login)
			user.GET("/:id", middleware.JWT(jwtSecret), h.GetUserInfo)	// 中间件校验 JWT 令牌
		}
	}
}

