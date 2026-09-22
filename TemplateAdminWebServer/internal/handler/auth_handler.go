// Package handler 提供 TemplateAdminWebServer 的 HTTP handlers。
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"template-mall/TemplateAdminWebServer/internal/auth"
	"template-mall/TemplateAdminWebServer/internal/config"
	"template-mall/TemplateAdminWebServer/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证 handler。
type AuthHandler struct {
	provider auth.AuthProvider
	store    *auth.SessionStore
	cfg      *config.Config
}

// NewAuthHandler 创建认证 handler。
func NewAuthHandler(provider auth.AuthProvider, store *auth.SessionStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{provider: provider, store: store, cfg: cfg}
}

// Login GET /auth/login
// 跳转到 OAuth 授权页（WPS）或 Mock 登录页。
func (h *AuthHandler) Login(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "generate state failed"})
		return
	}
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	redirectURL := h.provider.GetRedirectURL(state)
	c.Redirect(http.StatusFound, redirectURL)
}

// Callback GET /auth/callback
// OAuth 回调：校验 state → code 换 token → 拉用户信息 → 校验 ADMIN_USER_IDS → 写会话 → 跳转前端。
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "oauth error: " + errorParam})
		return
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "missing code"})
		return
	}

	// 校验 state
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState == "" || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid state"})
		return
	}
	// 清除 state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// code → token
	token, err := h.provider.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "exchange code failed: " + err.Error()})
		return
	}

	// token → userinfo
	userInfo, err := h.provider.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": "get userinfo failed: " + err.Error()})
		return
	}

	// 校验 user_id 是否在管理员白名单
	if !h.cfg.IsAdmin(userInfo.UserID) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": fmt.Sprintf("user %s (%s) not in admin whitelist, add to ADMIN_USER_IDS", userInfo.UserID, userInfo.Username),
		})
		return
	}

	// 写会话
	sid := h.store.Create(userInfo.UserID, userInfo.Username, userInfo.Avatar)
	c.SetCookie(middleware.SessionCookieName, sid, int(8*time.Hour/time.Second), "/", "", false, true)

	// 跳转前端首页
	c.Redirect(http.StatusFound, h.cfg.FrontendHomeURL)
}

// MockLogin POST /auth/mock-login
// AUTH_PROVIDER=mock 时使用：直接用 .env 固定账号登录。
func (h *AuthHandler) MockLogin(c *gin.Context) {
	if !h.cfg.MockAuthEnabled {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "mock auth disabled"})
		return
	}

	userInfo, err := h.provider.GetUserInfo(&auth.Token{AccessToken: "mock"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	if !h.cfg.IsAdmin(userInfo.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "user not in admin whitelist"})
		return
	}

	sid := h.store.Create(userInfo.UserID, userInfo.Username, userInfo.Avatar)
	c.SetCookie(middleware.SessionCookieName, sid, int(8*time.Hour/time.Second), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user_id":  userInfo.UserID,
			"username": userInfo.Username,
		},
	})
}

// Logout POST/GET /auth/logout （会话鉴权）
// 清会话。
func (h *AuthHandler) Logout(c *gin.Context) {
	sid, err := c.Cookie(middleware.SessionCookieName)
	if err == nil && sid != "" {
		h.store.Delete(sid)
	}
	// 清 Cookie
	c.SetCookie(middleware.SessionCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "logout success"})
}

// Me GET /auth/me （会话鉴权）
// 返回当前管理员。
func (h *AuthHandler) Me(c *gin.Context) {
	session, ok := middleware.GetSession(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "no session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user_id":  session.UserID,
			"username": session.Username,
			"avatar":   session.Avatar,
		},
	})
}

// Health GET /health 健康检查（无鉴权）。
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "TemplateAdminWebServer",
		"version": "1.0.0",
	})
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
