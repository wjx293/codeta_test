package repository

import (
	"log"
	"time"

	"template-mall/PayWebServer/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接。
var DB *gorm.DB

// InitDB 初始化 GORM 连接并配置连接池。
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  false,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("database connected")
	return nil
}

// AutoMigrate 自动迁移 pay_* 表结构（开发/演示环境）。
func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.PayPayment{},
		&model.PayCallback{},
		&model.PayOutbox{},
	)
}