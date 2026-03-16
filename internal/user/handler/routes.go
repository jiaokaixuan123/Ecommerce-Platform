package handler

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
			user.GET("/:id", middleware.JWT(jwtSecret), h.GetUserInfo)
		}
	}
}
