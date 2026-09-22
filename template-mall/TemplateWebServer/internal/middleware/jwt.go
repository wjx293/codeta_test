// Package middleware 提供 HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"template-mall/TemplateWebServer/internal/jwt"

	"github.com/gin-gonic/gin"
)

// ContextKeys 存放 context key 常量。
const (
	ContextKeyUserID       = "user_id"
	ContextKeyUsername     = "username"
	ContextKeyMemberStatus = "member_status"
)

// JWTAuth 返回 JWT 鉴权中间件。
// 从 Authorization Bearer 提取 access_token，校验签名和过期时间。
// 失败返回 401；成功将 user_id/username/member_status 注入 gin.Context。
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			return
		}

		// 期望格式：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "invalid authorization format, expected 'Bearer <token>'",
			})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "empty token",
			})
			return
		}

		claims, err := jwt.ParseAccessToken(secret, tokenStr)
		if err != nil {
			msg := "token invalid"
			if err == jwt.ErrTokenExpired {
				msg = "token expired"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": msg,
			})
			return
		}

		// 注入用户信息到 context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyMemberStatus, claims.MemberStatus)

		c.Next()
	}
}

// GetUserID 从 gin.Context 提取 user_id。
func GetUserID(c *gin.Context) (uint64, bool) {
	v, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	uid, ok := v.(uint64)
	return uid, ok
}

// GetUsername 从 gin.Context 提取 username。
func GetUsername(c *gin.Context) (string, bool) {
	v, exists := c.Get(ContextKeyUsername)
	if !exists {
		return "", false
	}
	name, ok := v.(string)
	return name, ok
}

// GetMemberStatus 从 gin.Context 提取 member_status。
func GetMemberStatus(c *gin.Context) (int8, bool) {
	v, exists := c.Get(ContextKeyMemberStatus)
	if !exists {
		return 0, false
	}
	status, ok := v.(int8)
	return status, ok
}
