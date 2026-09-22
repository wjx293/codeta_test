package repository

import (
	"errors"
	"template-mall/TemplateOrderServer/internal/model"

	"gorm.io/gorm"
)

var (
	ErrDuplicate    = errors.New("duplicate key violation")
	ErrNotFound     = errors.New("record not found")
	ErrNoRowsAffected = errors.New("no rows affected")
)

// UserRepo 用户仓储
type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

// Create 创建用户，捕获 uk_username 冲突。
func (r *UserRepo) Create(user *model.TosUser) error {
	err := r.db.Create(user).Error
	if err != nil && isDuplicate(err) {
		return ErrDuplicate
	}
	return err
}

// FindByUsername 按用户名查找。
func (r *UserRepo) FindByUsername(username string) (*model.TosUser, error) {
	var user model.TosUser
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &user, err
}

// FindByID 按 ID 查找。
func (r *UserRepo) FindByID(id uint64) (*model.TosUser, error) {
	var user model.TosUser
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &user, err
}

// UpdateMemberStatus 更新会员状态。
func (r *UserRepo) UpdateMemberStatus(userID uint64, status int8) error {
	result := r.db.Model(&model.TosUser{}).Where("id = ?", userID).Update("member_status", status)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

// ListUsers 分页查询用户，支持 username 模糊搜索。
func (r *UserRepo) ListUsers(username string, offset, limit int) ([]model.TosUser, int64, error) {
	var users []model.TosUser
	var total int64

	query := r.db.Model(&model.TosUser{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// isDuplicate 判断是否为 MySQL 唯一键冲突错误。
func isDuplicate(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		contains(err.Error(), "Duplicate entry") ||
		contains(err.Error(), "1062"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}