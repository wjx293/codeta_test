// Package consumer 提供 Kafka 消费者实现。
package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// PaymentEvent Kafka 支付事件消息
type PaymentEvent struct {
	MsgKey          string `json:"msg_key"`           // payment_no 或 callback_no
	BusinessOrderNo string `json:"business_order_no"`
	Amount          int64  `json:"amount"`
	Result          string `json:"result"` // success / fail
}

// Handler 消费处理函数接口
type Handler interface {
	HandlePaymentEvent(ctx context.Context, event *PaymentEvent) error
}

// KafkaConsumer Kafka 消费者
type KafkaConsumer struct {
	reader  *kafka.Reader
	handler Handler
}

// NewKafkaConsumer 创建 Kafka 消费者。
func NewKafkaConsumer(brokers []string, topic, groupID string, handler Handler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // 手动提交
		StartOffset:    kafka.LastOffset,
		MaxWait:        1 * time.Second,
	})

	return &KafkaConsumer{
		reader:  reader,
		handler: handler,
	}
}

// Start 启动消费循环，阻塞直到 ctx 取消。
func (c *KafkaConsumer) Start(ctx context.Context) error {
	log.Printf("kafka consumer started, topic=%s group=%s", c.reader.Config().Topic, c.reader.Config().GroupID)
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer shutting down")
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("kafka fetch error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var event PaymentEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("kafka unmarshal error: %v, offset=%d", err, msg.Offset)
			// 无法解析的消息也提交，避免阻塞
			if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
				log.Printf("kafka commit error: %v", commitErr)
			}
			continue
		}

		// 消息 Key 校验
		if event.MsgKey == "" {
			log.Printf("kafka message missing msg_key, offset=%d", msg.Offset)
			if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
				log.Printf("kafka commit error: %v", commitErr)
			}
			continue
		}

		// 业务处理失败不提交 offset，等待重试
		if err := c.handler.HandlePaymentEvent(ctx, &event); err != nil {
			log.Printf("kafka handle error: %v, msg_key=%s", err, event.MsgKey)
			continue
		}

		if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
			log.Printf("kafka commit error: %v", commitErr)
		}
	}
}