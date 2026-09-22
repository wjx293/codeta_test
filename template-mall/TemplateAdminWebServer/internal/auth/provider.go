// Package auth 提供管理员认证（WPS OAuth2 / Mock）。
package auth

// Token OAuth2 令牌。
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

// UserInfo 用户信息（从 OAuth Provider 拉取）。
type UserInfo struct {
	UserID   string // WPS user_id 或 Mock admin id
	Username string // 显示名
	Email    string
	Avatar   string
}

// AuthProvider 认证提供者抽象。
// 实现方：WPSOAuthProvider（生产）、MockAuthProvider（开发演示）。
// 通过 .env AUTH_PROVIDER=wps|mock 切换，预留公司 Auth Provider 实现点。
type AuthProvider interface {
	// GetRedirectURL 返回授权页跳转地址（含 state 防 CSRF）。
	GetRedirectURL(state string) string

	// ExchangeCode 用授权码换 access_token。
	ExchangeCode(code string) (*Token, error)

	// GetUserInfo 用 access_token 拉取用户信息。
	GetUserInfo(token *Token) (*UserInfo, error)
}
