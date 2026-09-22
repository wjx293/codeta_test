// Package config 提供 PayWebServer 的配置加载功能。
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config 包含 PayWebServer 的所有配置项。
type Config struct {
	// HTTP
	HTTPPort string

	// MySQL
	MySQLDSN string

	// Kafka
	KafkaBrokers []string
	KafkaTopic   string
}

// Load 从环境变量加载配置，缺失时使用默认值。
func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		HTTPPort:     env("HTTP_PORT", "8083"),
		MySQLDSN:     env("MYSQL_DSN", "root:root123456@tcp(127.0.0.1:3306)/template_order_db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"),
		KafkaBrokers: envSlice("KAFKA_BROKERS", "127.0.0.1:9092"),
		KafkaTopic:   env("KAFKA_TOPIC", "template-pay-events"),
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

// String 返回配置摘要。
func (c *Config) String() string {
	return fmt.Sprintf(
		"HTTPPort=%s MySQLDSN=%s KafkaBrokers=%v KafkaTopic=%s",
		c.HTTPPort, c.MySQLDSN, c.KafkaBrokers, c.KafkaTopic,
	)
}
