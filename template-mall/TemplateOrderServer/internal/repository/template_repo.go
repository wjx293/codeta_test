package repository

import (
	"time"

	"template-mall/TemplateOrderServer/internal/model"

	"gorm.io/gorm"
)

// TemplateRepo 模板仓储
type TemplateRepo struct{ db *gorm.DB }

func NewTemplateRepo(db *gorm.DB) *TemplateRepo { return &TemplateRepo{db: db} }

// Create 创建模板。
func (r *TemplateRepo) Create(template *model.TosTemplate) error {
	return r.db.Create(template).Error
}

// MaxTemplateNoToday 返回当日最大模板编号（无记录时返回空串）。
func (r *TemplateRepo) MaxTemplateNoToday() (string, error) {
	today := time.Now().Format("20060102")
	prefix := "TP" + today + "%"
	var maxNo *string
	err := r.db.Model(&model.TosTemplate{}).
		Where("template_no LIKE ?", prefix).
		Select("MAX(template_no)").
		Scan(&maxNo).Error
	if err != nil {
		return "", err
	}
	if maxNo == nil {
		return "", nil
	}
	return *maxNo, nil
}

// Update 更新模板；缩略图仅在 thumbnail_oss_key 非空时更新，避免编辑元数据时清空已有缩略图。
func (r *TemplateRepo) Update(template *model.TosTemplate) error {
	updates := map[string]interface{}{
		"name":        template.Name,
		"description": template.Description,
		"is_free":     template.IsFree,
		"price":       template.Price,
		"updated_at":  template.UpdatedAt,
	}
	if template.ThumbnailOssKey != "" {
		updates["thumbnail_oss_key"] = template.ThumbnailOssKey
		updates["thumbnail_file_size"] = template.ThumbnailFileSize
	}

	result := r.db.Model(&model.TosTemplate{}).
		Where("id = ? AND deleted_at IS NULL", template.ID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

// UpdateStatus 更新模板上下架状态。
func (r *TemplateRepo) UpdateStatus(id uint64, status int8) error {
	result := r.db.Model(&model.TosTemplate{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("status", status)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}

// FindByID 按 ID 查找（不包含已软删除的）。
func (r *TemplateRepo) FindByID(id uint64) (*model.TosTemplate, error) {
	var template model.TosTemplate
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &template, err
}

// FindByTemplateNo 按业务编号查找。
func (r *TemplateRepo) FindByTemplateNo(templateNo string) (*model.TosTemplate, error) {
	var template model.TosTemplate
	err := r.db.Where("template_no = ? AND deleted_at IS NULL", templateNo).First(&template).Error
	if err != nil {
		return nil, ErrNotFound
	}
	return &template, err
}

// ListTemplates 分页查询模板，支持按 status 筛选，不包含已软删除的。
func (r *TemplateRepo) ListTemplates(status *int8, offset, limit int) ([]model.TosTemplate, int64, error) {
	var templates []model.TosTemplate
	var total int64

	query := r.db.Model(&model.TosTemplate{}).Where("deleted_at IS NULL")
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// SoftDelete 软删除模板。
func (r *TemplateRepo) SoftDelete(id uint64) error {
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.TosTemplate{})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}