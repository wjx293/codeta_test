// 一次性迁移 pay_* 表，开发环境使用：go run ./scripts/migrate.go
package main

import (
	"log"

	"template-mall/PayWebServer/internal/config"
	"template-mall/PayWebServer/internal/repository"
)

func main() {
	cfg := config.Load()
	if err := repository.InitDB(cfg.MySQLDSN); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if err := repository.AutoMigrate(); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	log.Println("pay_* tables migrated successfully")
}
