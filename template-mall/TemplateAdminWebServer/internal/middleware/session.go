// Package middleware 提供 TemplateAdminWebServer 的中间件。
package middleware

import (
	"net/http"

	"template-mall/TemplateAdminWebServer/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	// SessionCookieName 会话 Cookie 名。
	SessionCookieName = "admin_sid"
	// ContextKeySession 会话在 gin.Context 的 key。
	ContextKeySession = "session"
)

// SessionAuth 会话鉴权中间件。
// 从 Cookie 读取 admin_sid → 校验会话存在且未过期 → 注入 *auth.Session 到 context。
// 失败返回 401。
func SessionAuth(store *auth.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie(SessionCookieName)
		if err != nil || sid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "missing session cookie",
			})
			return
		}

		session := store.Get(sid)
		if session == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "session expired or invalid",
			})
			return
		}

		c.Set(ContextKeySession, session)
		c.Next()
	}
}

// GetSession 从 context 获取当前会话。
func GetSession(c *gin.Context) (*auth.Session, bool) {
	v, ok := c.Get(ContextKeySession)
	if !ok {
		return nil, false
	}
	s, ok := v.(*auth.Session)
	return s, ok
}