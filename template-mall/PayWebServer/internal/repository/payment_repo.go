package repository

import (
	"errors"
	"strings"
	"time"

	"template-mall/PayWebServer/internal/model"

	"gorm.io/gorm"
)

var (
	ErrDuplicate = errors.New("duplicate")
	ErrNotFound  = errors.New("not found")
)

// PaymentRepoInterface 支付单仓储接口（用于测试 mock）
type PaymentRepoInterface interface {
	CreatePayment(p *model.PayPayment) error
	FindByBusinessOrderNo(orderNo string) (*model.PayPayment, error)
	FindByPaymentNo(paymentNo string) (*model.PayPayment, error)
	UpdateStatusConditionally(paymentNo string, fromStatus, toStatus int8) (int64, error)
}

// CallbackRepoInterface 回调记录仓储接口（用于测试 mock）
type CallbackRepoInterface interface {
	InsertCallback(c *model.PayCallback) error
}

// PaymentRepo 支付单仓储
type PaymentRepo struct{ db *gorm.DB }

func NewPaymentRepo(db *gorm.DB) *PaymentRepo { return &PaymentRepo{db: db} }

// CreatePayment 创建支付单。捕获 uk_business_order_no 冲突。
func (r *PaymentRepo) CreatePayment(p *model.PayPayment) error {
	err := r.db.Create(p).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// MaxPaymentNoToday 返回当日最大支付单编号（无记录时返回空串）。
func (r *PaymentRepo) MaxPaymentNoToday() (string, error) {
	today := time.Now().Format("20060102")
	prefix := "PW" + today + "%"
	var maxNo *string
	err := r.db.Model(&model.PayPayment{}).
		Where("payment_no LIKE ?", prefix).
		Select("MAX(payment_no)").
		Scan(&maxNo).Error
	if err != nil {
		return "", err
	}
	if maxNo == nil {
		return "", nil
	}
	return *maxNo, nil
}

// FindByBusinessOrderNo 根据业务订单号查支付单。
func (r *PaymentRepo) FindByBusinessOrderNo(orderNo string) (*model.PayPayment, error) {
	var p model.PayPayment
	err := r.db.Where("business_order_no = ?", orderNo).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// FindByPaymentNo 根据支付单号查支付单。
func (r *PaymentRepo) FindByPaymentNo(paymentNo string) (*model.PayPayment, error) {
	var p model.PayPayment
	err := r.db.Where("payment_no = ?", paymentNo).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// UpdateStatusConditionally 条件更新支付单状态，返回影响行数。
func (r *PaymentRepo) UpdateStatusConditionally(paymentNo string, fromStatus, toStatus int8) (int64, error) {
	result := r.db.Model(&model.PayPayment{}).
		Where("payment_no = ? AND status = ?", paymentNo, fromStatus).
		Update("status", toStatus)
	return result.RowsAffected, result.Error
}

// CallbackRepo 支付回调记录仓储
type CallbackRepo struct{ db *gorm.DB }

func NewCallbackRepo(db *gorm.DB) *CallbackRepo { return &CallbackRepo{db: db} }

// InsertCallback 插入回调记录，捕获 uk_callback_no 冲突。
func (r *CallbackRepo) InsertCallback(c *model.PayCallback) error {
	err := r.db.Create(c).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

func isDuplicate(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "1062"))
}
