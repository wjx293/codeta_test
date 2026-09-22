// Package config 提供 TemplateOrderServer 的配置加载功能。
// 配置来源：.env 文件 → 环境变量 → 默认值。
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config 包含 TemplateOrderServer 的所有配置项。
type Config struct {
	// gRPC
	GRPCPort string

	// MySQL
	MySQLDSN string

	// JWT
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	// Kafka
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	// PayWebServer
	PayWebServerURL string

	// OSS
	OSSEndpoint        string
	OSSAccessKeyID     string
	OSSAccessKeySecret string
	OSSBucket          string
	OSSRegion          string
}

// Load 从环境变量加载配置，缺失时使用默认值。
func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		GRPCPort:           env("GRPC_PORT", "9001"),
		MySQLDSN:           env("MYSQL_DSN", "root:root123456@tcp(127.0.0.1:3306)/template_order_db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"),
		JWTSecret:          env("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:       envDuration("JWT_ACCESS_TTL", 2*time.Hour),
		JWTRefreshTTL:      envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		KafkaBrokers:       envSlice("KAFKA_BROKERS", "127.0.0.1:9092"),
		KafkaTopic:         env("KAFKA_TOPIC", "template-pay-events"),
		KafkaGroupID:       env("KAFKA_GROUP_ID", "template-order-consumer-group"),
		PayWebServerURL:    env("PAY_WEB_SERVER_URL", "http://127.0.0.1:8083"),
		OSSEndpoint:        env("OSS_ENDPOINT", ""),
		OSSAccessKeyID:     env("OSS_ACCESS_KEY_ID", ""),
		OSSAccessKeySecret: env("OSS_ACCESS_KEY_SECRET", ""),
		OSSBucket:          env("OSS_BUCKET", ""),
		OSSRegion:          env("OSS_REGION", ""),
	}
	return cfg
}

func env(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return defaultVal
}

func envSlice(key, defaultVal string) []string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultVal
	}
	return []string{v}
}

// String 返回配置摘要（隐藏敏感字段）。
func (c *Config) String() string {
	return fmt.Sprintf(
		"GRPCPort=%s MySQLDSN=%s KafkaBrokers=%v PayWebServerURL=%s",
		c.GRPCPort, c.MySQLDSN, c.KafkaBrokers, c.PayWebServerURL,
	)
}
