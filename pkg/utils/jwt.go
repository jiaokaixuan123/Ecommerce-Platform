package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷结构体
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户 ID、角色、密钥和过期时间生成符合 JWT 规范的 Token
func GenerateToken(userID uint, role, secret string, expireHour int) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间：当前时间 + 过期小时数
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			// 签发时间：当前时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 创建 JWT Token：指定签名算法（HS256）+ 载荷
	// 用密钥签名 Token 并返回字符串
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenStr, secret string) (*Claims, error) {
	// 解析 Token：绑定自定义 Claims 结构体，指定验签回调函数
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		// 验签回调：返回签名密钥（注意类型转换为 []byte）
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
