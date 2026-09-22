package repository

import (
	"template-mall/PayWebServer/internal/model"

	"gorm.io/gorm"
)

// OutboxRepoInterface Outbox 仓储接口（用于测试 mock）
type OutboxRepoInterface interface {
	Insert(o *model.PayOutbox) error
	FindPending(limit int) ([]model.PayOutbox, error)
	MarkSent(eventKey string) error
}

// OutboxRepo Outbox 仓储
type OutboxRepo struct{ db *gorm.DB }

func NewOutboxRepo(db *gorm.DB) *OutboxRepo { return &OutboxRepo{db: db} }

// Insert 插入待发送 Outbox 记录。
func (r *OutboxRepo) Insert(o *model.PayOutbox) error {
	err := r.db.Create(o).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// FindPending 查询待发送记录（status=0），按创建时间升序。
func (r *OutboxRepo) FindPending(limit int) ([]model.PayOutbox, error) {
	var records []model.PayOutbox
	err := r.db.Where("status = ?", 0).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// MarkSent 将 Outbox 记录标记为已发送。
func (r *OutboxRepo) MarkSent(eventKey string) error {
	return r.db.Model(&model.PayOutbox{}).
		Where("event_key = ? AND status = ?", eventKey, 0).
		Updates(map[string]interface{}{
			"status":     1,
			"updated_at": gorm.Expr("NOW(3)"),
		}).Error
}
