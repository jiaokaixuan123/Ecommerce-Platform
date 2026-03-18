package middleware

import (
	"strings"

	"github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/ecommerce-platform/pkg/utils"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"   // 存储用户 ID 的键
const UserRoleKey = "user_role" // 存储用户角色的键

// JWT 生成 Gin 中间件：接收 JWT 签名密钥，返回认证中间件函数
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 Authorization 字段
		auth := c.GetHeader("Authorization")
		// 必须为 Bearer 开头，否则未授权
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, errors.ErrUnauthorized, errors.Msg(errors.ErrUnauthorized))
			c.Abort()	// 终止请求链（不再执行后续中间件/处理器）
			return
		}

		// 提取 Token 并解析
		claims, err := utils.ParseToken(strings.TrimPrefix(auth, "Bearer "), secret)
		if err != nil {
			response.Fail(c, errors.ErrTokenInvalid, errors.Msg(errors.ErrTokenInvalid))
			c.Abort()
			return
		}

		// 解析成功：将用户信息存入 Gin 上下文
		c.Set(UserIDKey, claims.UserID)   // 存储用户 ID
		c.Set(UserRoleKey, claims.Role)   // 存储用户角色
		c.Next() 						  // 执行后续中间件/处理器
	}
}

// GetUserID 从 Gin 上下文获取用户 ID
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(UserIDKey)
	uid, _ := id.(uint)
	return uid
}

// RoleAuth 角色权限中间件，需在 JWT 中间件之后使用
// 传入允许访问的角色列表，不在列表中的角色返回 403
func RoleAuth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(UserRoleKey)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		response.Fail(c, errors.ErrForbidden, errors.Msg(errors.ErrForbidden))
		c.Abort()
	}
}
