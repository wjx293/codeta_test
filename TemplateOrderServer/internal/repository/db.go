// Package repository 提供 GORM 数据访问层。
package repository

import (
	"log"
	"time"

	"template-mall/TemplateOrderServer/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例。
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

// AutoMigrate 自动迁移表结构（开发环境），与 migrations DDL 保持一致。
func AutoMigrate() error {
	if err := dedupeMonthlyOrders(); err != nil {
		return err
	}
	if err := DB.AutoMigrate(
		&model.TosUser{},
		&model.TosRefreshToken{},
		&model.TosTemplate{},
		&model.TosOrder{},
		&model.TosKafkaConsumeRecord{},
	); err != nil {
		return err
	}
	return EnsureOrderConstraints()
}

// dedupeMonthlyOrders 删除同用户×模板×月份×类型的重复免费/会员订单，保留最早一条。
func dedupeMonthlyOrders() error {
	return DB.Exec(`
		DELETE o1 FROM tos_orders o1
		INNER JOIN tos_orders o2
		WHERE o1.user_id = o2.user_id
		  AND o1.template_id = o2.template_id
		  AND o1.month = o2.month
		  AND o1.order_type = o2.order_type
		  AND o1.month IS NOT NULL
		  AND o1.id > o2.id
	`).Error
}

// EnsureOrderConstraints 补齐 AutoMigrate 无法创建的生成列与唯一索引。
func EnsureOrderConstraints() error {
	var colCount int64
	if err := DB.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'tos_orders' AND column_name = 'active_unpaid_key'
	`).Scan(&colCount).Error; err != nil {
		return err
	}
	if colCount == 0 {
		if err := DB.Exec(`
			ALTER TABLE tos_orders
			ADD COLUMN active_unpaid_key VARCHAR(64) AS (
				CASE WHEN order_type = 3 AND status = 1
					 THEN CONCAT(user_id, '_', template_id)
					 ELSE NULL
				END
			) STORED
		`).Error; err != nil {
			return err
		}
	}

	if err := ensureIndex("tos_orders", "uk_active_unpaid", "ADD UNIQUE KEY uk_active_unpaid (active_unpaid_key)"); err != nil {
		return err
	}
	if err := ensureIndex("tos_orders", "uk_free_month", "ADD UNIQUE KEY uk_free_month (user_id, template_id, month, order_type)"); err != nil {
		return err
	}
	return ensureIndex("tos_orders", "uk_member_month", "ADD UNIQUE KEY uk_member_month (user_id, template_id, month, order_type)")
}

func ensureIndex(table, indexName, ddl string) error {
	var count int64
	if err := DB.Raw(`
		SELECT COUNT(*) FROM information_schema.statistics
		WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?
	`, table, indexName).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return DB.Exec("ALTER TABLE " + table + " " + ddl).Error
}
