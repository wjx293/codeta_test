package service

import (
	"context"
	"log"
	"time"

	"template-mall/TemplateOrderServer/internal/consumer"
	"template-mall/TemplateOrderServer/internal/model"
	"template-mall/TemplateOrderServer/internal/repository"

	"gorm.io/gorm"
)

// ConsumeService 消费支付事件、更新订单状态
type ConsumeService struct {
	db *gorm.DB
}

// NewConsumeService 创建消费服务实例。
func NewConsumeService(db *gorm.DB) *ConsumeService {
	return &ConsumeService{db: db}
}

// HandlePaymentEvent 实现 consumer.Handler 接口：处理支付回调。
func (s *ConsumeService) HandlePaymentEvent(ctx context.Context, event *consumer.PaymentEvent) error {
	if event.Result != "success" {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		kafkaRepo := repository.NewKafkaRepo(tx)
		orderRepo := repository.NewOrderRepo(tx)

		record := &model.TosKafkaConsumeRecord{
			MsgKey:    event.MsgKey,
			CreatedAt: time.Now(),
		}
		if err := kafkaRepo.InsertConsumeRecord(record); err != nil {
			if err == repository.ErrDuplicate {
				// 幂等记录已存在：补试订单更新（修复历史部分成功场景）
				_, updateErr := orderRepo.UpdateStatusConditionally(event.BusinessOrderNo, 1, 2)
				return updateErr
			}
			return err
		}

		rows, err := orderRepo.UpdateStatusConditionally(event.BusinessOrderNo, 1, 2)
		if err != nil {
			return err
		}
		if rows == 0 {
			log.Printf("order status update no rows affected, order_no=%s", event.BusinessOrderNo)
		} else {
			log.Printf("order status updated to paid, order_no=%s", event.BusinessOrderNo)
		}
		return nil
	})
}
