package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"template-mall/PayWebServer/internal/model"
	"template-mall/PayWebServer/internal/producer"
	"template-mall/PayWebServer/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrAmountMismatch  = errors.New("amount mismatch")
	ErrStatusInvalid   = errors.New("payment status invalid")
	ErrKafkaProduce    = errors.New("kafka produce failed")
)

// PaymentService 支付单服务
type PaymentService struct {
	paymentRepo repository.PaymentRepoInterface
	cbRepo      repository.CallbackRepoInterface
	outboxRepo  repository.OutboxRepoInterface
	producer    *producer.KafkaProducer
	db          *gorm.DB
}

// NewPaymentService 创建支付单服务实例。
func NewPaymentService(
	paymentRepo repository.PaymentRepoInterface,
	cbRepo repository.CallbackRepoInterface,
	outboxRepo repository.OutboxRepoInterface,
	p *producer.KafkaProducer,
	db *gorm.DB,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		cbRepo:      cbRepo,
		outboxRepo:  outboxRepo,
		producer:    p,
		db:          db,
	}
}

// CreatePaymentResponse 创建支付单响应
type CreatePaymentResponse struct {
	PaymentNo       string
	BusinessOrderNo string
	Amount          int64
	Status          int8
}

// CreatePayment 创建支付单（幂等：相同 business_order_no 返回相同 payment_no）。
func (s *PaymentService) CreatePayment(businessOrderNo string, amount int64, paymentNo string) (*CreatePaymentResponse, error) {
	now := time.Now()

	payment := &model.PayPayment{
		PaymentNo:       paymentNo,
		BusinessOrderNo: businessOrderNo,
		Amount:          amount,
		Status:          1, // 未支付
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.paymentRepo.CreatePayment(payment); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			existing, findErr := s.paymentRepo.FindByBusinessOrderNo(businessOrderNo)
			if findErr != nil {
				return nil, findErr
			}
			return &CreatePaymentResponse{
				PaymentNo:       existing.PaymentNo,
				BusinessOrderNo: existing.BusinessOrderNo,
				Amount:          existing.Amount,
				Status:          existing.Status,
			}, nil
		}
		return nil, err
	}

	return &CreatePaymentResponse{
		PaymentNo:       payment.PaymentNo,
		BusinessOrderNo: payment.BusinessOrderNo,
		Amount:          payment.Amount,
		Status:          payment.Status,
	}, nil
}

// ProcessCallbackResponse 回调处理结果
type ProcessCallbackResponse struct {
	PaymentNo       string
	BusinessOrderNo string
	Amount          int64
	Result          string
	IsNew           bool
}

// HandleCallback 处理支付回调。
// 流程：幂等插入回调 → 校验支付单 →（仅 success）事务更新状态+Outbox → Kafka 生产。
func (s *PaymentService) HandleCallback(ctx context.Context, callbackNo, paymentNo, result string, amount int64) (*ProcessCallbackResponse, error) {
	payment, err := s.paymentRepo.FindByPaymentNo(paymentNo)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	now := time.Now()
	cb := &model.PayCallback{
		CallbackNo: callbackNo,
		PaymentNo:  paymentNo,
		Amount:     amount,
		Result:     result,
		CreatedAt:  now,
	}

	if err := s.cbRepo.InsertCallback(cb); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return s.handleDuplicateCallback(ctx, payment, callbackNo, paymentNo, result, amount)
		}
		return nil, err
	}

	if result != "success" {
		return &ProcessCallbackResponse{
			PaymentNo:       paymentNo,
			BusinessOrderNo: payment.BusinessOrderNo,
			Amount:          amount,
			Result:          result,
			IsNew:           true,
		}, nil
	}

	if payment.Status != 1 {
		return nil, ErrStatusInvalid
	}
	if payment.Amount != amount {
		return nil, ErrAmountMismatch
	}

	if err := s.completeSuccessPayment(ctx, payment, callbackNo, amount); err != nil {
		return nil, err
	}

	return &ProcessCallbackResponse{
		PaymentNo:       paymentNo,
		BusinessOrderNo: payment.BusinessOrderNo,
		Amount:          amount,
		Result:          result,
		IsNew:           true,
	}, nil
}

// handleDuplicateCallback 处理重复回调；success 且支付单仍未支付时尝试补全状态与事件。
func (s *PaymentService) handleDuplicateCallback(
	ctx context.Context,
	payment *model.PayPayment,
	callbackNo, paymentNo, result string,
	amount int64,
) (*ProcessCallbackResponse, error) {
	resp := &ProcessCallbackResponse{
		PaymentNo:       paymentNo,
		BusinessOrderNo: payment.BusinessOrderNo,
		Amount:          amount,
		Result:          result,
		IsNew:           false,
	}

	if result != "success" {
		return resp, nil
	}

	if payment.Amount != amount {
		return nil, ErrAmountMismatch
	}
	if payment.Status == 2 {
		return resp, nil
	}
	if payment.Status != 1 {
		return nil, ErrStatusInvalid
	}

	if err := s.completeSuccessPayment(ctx, payment, callbackNo, amount); err != nil {
		return nil, err
	}
	return resp, nil
}

// completeSuccessPayment 事务更新支付单状态并写入 Outbox，随后发送 Kafka。
func (s *PaymentService) completeSuccessPayment(
	ctx context.Context,
	payment *model.PayPayment,
	callbackNo string,
	amount int64,
) error {
	event := &producer.PaymentEvent{
		MsgKey:          callbackNo,
		PaymentNo:       payment.PaymentNo,
		BusinessOrderNo: payment.BusinessOrderNo,
		Amount:          amount,
		Result:          "success",
		CallbackNo:      callbackNo,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	now := time.Now()
	updated, err := s.persistSuccessPayment(payment.PaymentNo, callbackNo, string(payload), now)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	if s.producer == nil {
		return nil
	}

	if err := s.producer.ProducePaymentEvent(ctx, event); err != nil {
		return fmt.Errorf("%w: %v", ErrKafkaProduce, err)
	}

	if s.outboxRepo != nil {
		_ = s.outboxRepo.MarkSent(callbackNo)
	}
	return nil
}

// persistSuccessPayment 在同一事务中更新支付单状态并写入 Outbox。
func (s *PaymentService) persistSuccessPayment(paymentNo, eventKey, payload string, now time.Time) (bool, error) {
	if s.db != nil {
		var updated bool
		err := s.db.Transaction(func(tx *gorm.DB) error {
			ok, err := s.persistSuccessPaymentTx(tx, paymentNo, eventKey, payload, now)
			updated = ok
			return err
		})
		return updated, err
	}

	rows, err := s.paymentRepo.UpdateStatusConditionally(paymentNo, 1, 2)
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, nil
	}
	if s.outboxRepo != nil {
		if err := s.outboxRepo.Insert(&model.PayOutbox{
			EventKey:  eventKey,
			Payload:   payload,
			Status:    0,
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil && !errors.Is(err, repository.ErrDuplicate) {
			return false, err
		}
	}
	return true, nil
}

func (s *PaymentService) persistSuccessPaymentTx(tx *gorm.DB, paymentNo, eventKey, payload string, now time.Time) (bool, error) {
	paymentRepo := repository.NewPaymentRepo(tx)
	outboxRepo := repository.NewOutboxRepo(tx)

	rows, err := paymentRepo.UpdateStatusConditionally(paymentNo, 1, 2)
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, nil
	}

	if err := outboxRepo.Insert(&model.PayOutbox{
		EventKey:  eventKey,
		Payload:   payload,
		Status:    0,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil && !errors.Is(err, repository.ErrDuplicate) {
		return false, err
	}
	return true, nil
}

// RelayOutbox 补发待发送的 Outbox 事件（供后台 worker 调用）。
func (s *PaymentService) RelayOutbox(ctx context.Context, limit int) error {
	if s.outboxRepo == nil || s.producer == nil {
		return nil
	}

	records, err := s.outboxRepo.FindPending(limit)
	if err != nil {
		return err
	}

	for _, rec := range records {
		var event producer.PaymentEvent
		if err := json.Unmarshal([]byte(rec.Payload), &event); err != nil {
			continue
		}
		if err := s.producer.ProducePaymentEvent(ctx, &event); err != nil {
			return fmt.Errorf("%w: %v", ErrKafkaProduce, err)
		}
		if err := s.outboxRepo.MarkSent(rec.EventKey); err != nil {
			return err
		}
	}
	return nil
}
