package repository

import (
	"template-mall/TemplateOrderServer/internal/model"
	"time"

	"gorm.io/gorm"
)

// RefreshTokenRepo Refresh Token 仓储
type RefreshTokenRepo struct{ db *gorm.DB }

func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo { return &RefreshTokenRepo{db: db} }

// Create 创建 refresh token 记录。
func (r *RefreshTokenRepo) Create(token *model.TosRefreshToken) error {
	err := r.db.Create(token).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// FindByTokenHash 按 token_hash 查找。
func (r *RefreshTokenRepo) FindByTokenHash(tokenHash string) (*model.TosRefreshToken, error) {
	var token model.TosRefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &token, nil
}

// RevokeByUserID 吊销用户所有未吊销的 refresh token。
func (r *RefreshTokenRepo) RevokeByUserID(userID uint64) error {
	return r.db.Model(&model.TosRefreshToken{}).
		Where("user_id = ? AND revoked = 0", userID).
		Update("revoked", 1).Error
}

// RevokeByTokenHash 吊销单个 refresh token。
func (r *RefreshTokenRepo) RevokeByTokenHash(tokenHash string) error {
	result := r.db.Model(&model.TosRefreshToken{}).
		Where("token_hash = ? AND revoked = 0", tokenHash).
		Update("revoked", 1)
	if result.RowsAffected == 0 {
		return ErrNoRowsAffected
	}
	return result.Error
}

// DeleteExpired 删除过期的 token 记录。
func (r *RefreshTokenRepo) DeleteExpired() (int64, error) {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&model.TosRefreshToken{})
	return result.RowsAffected, result.Error
}