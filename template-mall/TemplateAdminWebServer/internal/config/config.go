// Package config 提供 TemplateAdminWebServer 的配置加载功能。
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config 包含 TemplateAdminWebServer 的所有配置项。
type Config struct {
	// HTTP
	HTTPPort string

	// gRPC
	GRPCTarget string

	// Auth Provider：wps | mock
	AuthProvider string

	// WPS OAuth2
	WPSClientID       string
	WPSClientSecret   string
	WPSRedirectURI    string
	WPSScope          string
	WPSAuthURL        string
	WPSTokenURL       string
	WPSUserInfoURL    string
	WPSKSOSignEnabled bool

	// 管理员白名单（逗号分隔的 user_id）
	AdminUserIDs []string

	// 前端首页（OAuth 回调成功后跳转）
	FrontendHomeURL string

	// Mock Auth（AUTH_PROVIDER=mock 时使用）
	MockAuthEnabled bool
	MockAdminUserID string
	MockAdminName   string

	// OSS（STS + HeadObject）
	OSSEndpoint        string
	OSSAccessKeyID     string
	OSSAccessKeySecret string
	OSSBucket          string
	OSSRegion          string

	// STS Role（生产环境用）
	STSRoleARN     string
	STSSessionName string
}

// Load 从环境变量加载配置。
func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		HTTPPort:   env("HTTP_PORT", "8081"),
		GRPCTarget: env("GRPC_TARGET", "localhost:9001"),

		AuthProvider: env("AUTH_PROVIDER", "mock"),

		WPSClientID:       env("WPS_CLIENT_ID", "AK20230207WTGKYW"),
		WPSClientSecret:   env("WPS_CLIENT_SECRET", ""),
		WPSRedirectURI:    env("WPS_REDIRECT_URI", "http://localhost:3001/api/auth/callback"),
		WPSScope:          env("WPS_SCOPE", "kso.user_base.read"),
		WPSAuthURL:        env("WPS_AUTH_URL", "https://openapi.wps.cn/oauth2/auth"),
		WPSTokenURL:       env("WPS_TOKEN_URL", "https://openapi.wps.cn/oauth2/token"),
		WPSUserInfoURL:    env("WPS_USERINFO_URL", "https://openapi.wps.cn/v7/users/current"),
		WPSKSOSignEnabled: env("WPS_KSO_SIGN_ENABLED", "false") == "true",

		AdminUserIDs: parseList(env("ADMIN_USER_IDS", "10001,10002")),

		FrontendHomeURL: env("FRONTEND_HOME_URL", "http://localhost:3001/"),

		MockAuthEnabled: env("AUTH_PROVIDER", "mock") == "mock",
		MockAdminUserID: env("MOCK_ADMIN_USER_ID", "10001"),
		MockAdminName:   env("MOCK_ADMIN_NAME", "MockAdmin"),

		OSSEndpoint:        env("OSS_ENDPOINT", ""),
		OSSAccessKeyID:     env("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: env("OSS_ACCESS_KEY_SECRET", ""),
		OSSBucket:          env("OSS_BUCKET", ""),
		OSSRegion:          env("OSS_REGION", ""),

		STSRoleARN:     env("STS_ROLE_ARN", ""),
		STSSessionName: env("STS_SESSION_NAME", "template-mall-admin"),
	}
	return cfg
}

// IsAdmin 判断 user_id 是否在管理员白名单。
func (c *Config) IsAdmin(userID string) bool {
	for _, id := range c.AdminUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// String 返回配置摘要（不含敏感字段）。
func (c *Config) String() string {
	return fmt.Sprintf(
		"HTTPPort=%s GRPCTarget=%s AuthProvider=%s FrontendHomeURL=%s OSSBucket=%s",
		c.HTTPPort, c.GRPCTarget, c.AuthProvider, c.FrontendHomeURL, c.OSSBucket,
	)
}
