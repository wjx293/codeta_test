// Package config 提供 TemplateWebServer 的配置加载功能。
package config

import (
	"fmt"
	"os"
)

// Config 包含 TemplateWebServer 的所有配置项。
type Config struct {
	// HTTP
	HTTPPort string

	// gRPC
	GRPCTarget string

	// JWT
	JWTSecret string
}

// Load 从环境变量加载配置。
func Load() *Config {
	return &Config{
		HTTPPort:   env("HTTP_PORT", "8080"),
		GRPCTarget: env("GRPC_TARGET", "localhost:9001"),
		JWTSecret:  env("JWT_SECRET", "template-mall-secret-key"),
	}
}

func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// String 返回配置摘要。
func (c *Config) String() string {
	return fmt.Sprintf(
		"HTTPPort=%s GRPCTarget=%s",
		c.HTTPPort, c.GRPCTarget,
	)
}