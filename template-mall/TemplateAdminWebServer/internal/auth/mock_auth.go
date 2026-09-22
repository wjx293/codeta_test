package auth

import (
	"fmt"

	"template-mall/TemplateAdminWebServer/internal/config"
)

// MockAuthProvider 开发演示环境的 Mock 认证。
// AUTH_PROVIDER=mock 时使用，跳过 WPS，用 .env 固定账号校验。
// 仍走 AuthProvider 接口，方便切换。
type MockAuthProvider struct {
	cfg *config.Config
}

// NewMockAuthProvider 创建 Mock Provider。
func NewMockAuthProvider(cfg *config.Config) *MockAuthProvider {
	return &MockAuthProvider{cfg: cfg}
}

// GetRedirectURL Mock 模式不跳转 WPS，前端直接 POST /auth/login?mock=true。
// 这里返回 BFF 自己的 mock 登录端点。
func (p *MockAuthProvider) GetRedirectURL(state string) string {
	return fmt.Sprintf("/auth/mock-login?state=%s", state)
}

// ExchangeCode Mock 模式：code 即 mock user_id，直接返回固定 token。
func (p *MockAuthProvider) ExchangeCode(code string) (*Token, error) {
	return &Token{
		AccessToken:  "mock-access-token-" + code,
		RefreshToken: "mock-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
	}, nil
}

// GetUserInfo Mock 模式：返回 .env 配置的固定管理员账号。
func (p *MockAuthProvider) GetUserInfo(token *Token) (*UserInfo, error) {
	return &UserInfo{
		UserID:   p.cfg.MockAdminUserID,
		Username: p.cfg.MockAdminName,
	}, nil
}
