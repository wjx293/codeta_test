// Package producer 提供 Kafka 生产功能。
package producer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// PaymentEvent Kafka 支付事件消息
type PaymentEvent struct {
	MsgKey          string `json:"msg_key"`
	PaymentNo       string `json:"payment_no"`
	BusinessOrderNo string `json:"business_order_no"`
	Amount          int64  `json:"amount"`
	Result          string `json:"result"`
	CallbackNo      string `json:"callback_no"`
}

// KafkaProducer Kafka 生产者
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer 创建 Kafka 生产者。
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &KafkaProducer{writer: writer}
}

// ProducePaymentEvent 发送支付事件到 Kafka。
// key=business_order_no, value=PaymentEvent JSON.
func (p *KafkaProducer) ProducePaymentEvent(ctx context.Context, event *PaymentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.BusinessOrderNo),
		Value: body,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return err
	}

	log.Printf("kafka produced: key=%s msg_key=%s result=%s", event.BusinessOrderNo, event.MsgKey, event.Result)
	return nil
}

// Close 关闭生产者。
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}