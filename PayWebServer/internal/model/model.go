package model

import "time"

// PayPayment 支付单表（对应 pay_payments）
type PayPayment struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	PaymentNo       string    `gorm:"type:varchar(32);uniqueIndex:uk_payment_no;not null"`
	BusinessOrderNo string    `gorm:"type:varchar(32);uniqueIndex:uk_business_order_no;not null"`
	Amount          int64     `gorm:"not null"`
	Status          int8      `gorm:"type:tinyint;not null"` // 1=未支付 2=已支付
	CreatedAt       time.Time `gorm:"type:datetime(3);not null"`
	UpdatedAt       time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (PayPayment) TableName() string { return "pay_payments" }

// PayCallback 支付回调流水表（对应 pay_callbacks）
type PayCallback struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	CallbackNo string    `gorm:"type:varchar(32);uniqueIndex:uk_callback_no;not null"`
	PaymentNo  string    `gorm:"type:varchar(32);not null;index:idx_payment_no"`
	Amount     int64     `gorm:"not null"`
	Result     string    `gorm:"type:varchar(16);not null"` // success / fail
	CreatedAt  time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (PayCallback) TableName() string { return "pay_callbacks" }

// PayOutbox 支付事件 Outbox（对应 pay_outbox）
type PayOutbox struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	EventKey  string    `gorm:"type:varchar(64);uniqueIndex:uk_event_key;not null"`
	Payload   string    `gorm:"type:text;not null"`
	Status    int8      `gorm:"type:tinyint;not null;default:0"` // 0=待发送 1=已发送
	CreatedAt time.Time `gorm:"type:datetime(3);not null"`
	UpdatedAt time.Time `gorm:"type:datetime(3);not null"`
}

// TableName 指定表名。
func (PayOutbox) TableName() string { return "pay_outbox" }