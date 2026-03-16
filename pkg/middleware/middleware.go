package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"	// 令牌桶限流库
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始的时间
		start := time.Now()
		// 执行后续中间件/处理器
		c.Next()
		// 请求处理完成后，输出日志（结构化）
		zap.L().Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// RateLimiter 限流中间件（令牌桶）
// r：rate.Limit 类型，表示每秒生成的令牌数（请求频率上限）
// b：int 类型，表示令牌桶的容量（最大突发请求数）
func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	// 创建令牌桶限流器
	limiter := rate.NewLimiter(r, b)
	return func(c *gin.Context) {
		// 尝试获取令牌
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "too many requests"})
			c.Abort()
			return
		}
		// 获取令牌成功：执行后续中间件/处理器
		c.Next()
	}
}
