package service

import (
	"context"
	"log"
	"time"
)

// OutboxRelay 后台补发 Outbox 中待发送的 Kafka 事件。
type OutboxRelay struct {
	svc      *PaymentService
	interval time.Duration
}

// NewOutboxRelay 创建 Outbox relay。
func NewOutboxRelay(svc *PaymentService, interval time.Duration) *OutboxRelay {
	return &OutboxRelay{svc: svc, interval: interval}
}

// Start 启动 relay 循环，阻塞直到 ctx 取消。
func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	log.Printf("outbox relay started, interval=%s", r.interval)
	for {
		select {
		case <-ctx.Done():
			log.Println("outbox relay stopped")
			return
		case <-ticker.C:
			if err := r.svc.RelayOutbox(ctx, 50); err != nil {
				log.Printf("outbox relay error: %v", err)
			}
		}
	}
}
