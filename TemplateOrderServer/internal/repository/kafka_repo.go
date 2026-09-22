package repository

import (
	"template-mall/TemplateOrderServer/internal/model"

	"gorm.io/gorm"
)

// KafkaRepo Kafka 消费记录仓储
type KafkaRepo struct{ db *gorm.DB }

func NewKafkaRepo(db *gorm.DB) *KafkaRepo { return &KafkaRepo{db: db} }

// InsertConsumeRecord 插入消费记录，捕获 uk_msg_key 冲突。
func (r *KafkaRepo) InsertConsumeRecord(record *model.TosKafkaConsumeRecord) error {
	err := r.db.Create(record).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}