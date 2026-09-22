// Package model 定义 GORM 数据模型，与 migrations DDL 一致。
package model

import (
	"time"
)

// TosUser C 端用户表
type TosUser struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"type:varchar(64);uniqueIndex:uk_username;not null"`
	Password     string    `gorm:"type:varchar(255);not null"` // bcrypt
	MemberStatus int8      `gorm:"type:tinyint;not null;default:0"`
	CreatedAt    time.Time `gorm:"type:datetime(3);not null"`
	UpdatedAt    time.Time `gorm:"type:datetime(3);not null"`
}

func (TosUser) TableName() string { return "tos_users" }

// TosRefreshToken Refresh Token 记录表
type TosRefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"not null;index:idx_user_id"`
	TokenHash string    `gorm:"type:varchar(64);uniqueIndex:uk_token_hash;not null"`
	ExpiresAt time.Time `gorm:"type:datetime(3);not null;index:idx_expires_at"`
	Revoked   int8      `gorm:"type:tinyint;not null;default:0"`
	CreatedAt time.Time `gorm:"type:datetime(3);not null"`
}

func (TosRefreshToken) TableName() string { return "tos_refresh_tokens" }

// TosTemplate 模板元数据表
type TosTemplate struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	TemplateNo        string     `gorm:"type:varchar(32);uniqueIndex:uk_template_no;not null"`
	Name              string     `gorm:"type:varchar(128);not null"`
	Description       string     `gorm:"type:text"`
	IsFree            int8       `gorm:"type:tinyint;not null"`
	Price             int64      `gorm:"not null"`
	Status            int8       `gorm:"type:tinyint;not null"`
	OSSType           string     `gorm:"type:varchar(16);not null"`
	BucketName        string     `gorm:"type:varchar(64);not null"`
	OriginalFilename  string     `gorm:"type:varchar(255);not null"`
	FileType          string     `gorm:"type:varchar(8);not null"`
	FileSize          uint64     `gorm:"not null"`
	FileOssKey        string     `gorm:"type:varchar(255);not null"`
	ThumbnailOssKey   string     `gorm:"type:varchar(255)"`
	ThumbnailFileSize uint64     `gorm:""`
	CreatedAt         time.Time  `gorm:"type:datetime(3);not null"`
	UpdatedAt         time.Time  `gorm:"type:datetime(3);not null"`
	DeletedAt         *time.Time `gorm:"type:datetime(3);index:idx_deleted_at"`
}

func (TosTemplate) TableName() string { return "tos_templates" }

// TosOrder 模板订单表
type TosOrder struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	OrderNo       string    `gorm:"type:varchar(32);uniqueIndex:uk_order_no;not null"`
	UserID        uint64    `gorm:"not null;index:idx_user_status,priority:1;uniqueIndex:uk_free_month,priority:1;uniqueIndex:uk_member_month,priority:1"`
	TemplateID    uint64    `gorm:"not null;index:idx_template;uniqueIndex:uk_free_month,priority:2;uniqueIndex:uk_member_month,priority:2"`
	OrderType     int8      `gorm:"type:tinyint;not null;uniqueIndex:uk_free_month,priority:4;uniqueIndex:uk_member_month,priority:4"` // 1=免费 2=会员 3=零售
	PriceSnapshot int64     `gorm:"not null"`
	Status        int8      `gorm:"type:tinyint;not null;index:idx_user_status,priority:2"` // 1=未支付 2=已获得下载资格 3=已取消
	Month         *string   `gorm:"type:char(7);uniqueIndex:uk_free_month,priority:3;uniqueIndex:uk_member_month,priority:3"` // 免费/会员填 yyyyMM，零售 NULL
	CreatedAt     time.Time `gorm:"type:datetime(3);not null"`
	UpdatedAt     time.Time `gorm:"type:datetime(3);not null"`
}

func (TosOrder) TableName() string { return "tos_orders" }

// TosKafkaConsumeRecord Kafka 消费幂等记录表
type TosKafkaConsumeRecord struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	MsgKey          string    `gorm:"type:varchar(64);uniqueIndex:uk_msg_key;not null"`
	BusinessOrderNo string    `gorm:"type:varchar(32);not null"`
	Result          string    `gorm:"type:varchar(16);not null"`
	CreatedAt       time.Time `gorm:"type:datetime(3);not null"`
}

func (TosKafkaConsumeRecord) TableName() string { return "tos_kafka_consume_records" }

// OrderView 订单视图（JOIN 查询结果）
type OrderView struct {
	ID            uint64
	OrderNo       string
	UserID        uint64
	Username      string
	TemplateID    uint64
	TemplateName  string
	OrderType     int8
	PriceSnapshot int64
	Status        int8
	Month         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
