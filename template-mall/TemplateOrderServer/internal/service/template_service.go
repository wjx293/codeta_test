package service

import (
	"errors"
	"template-mall/TemplateOrderServer/internal/idgen"
	"template-mall/TemplateOrderServer/internal/model"
	"template-mall/TemplateOrderServer/internal/oss"
	"template-mall/TemplateOrderServer/internal/repository"
	"time"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrPriceInvalid     = errors.New("free template must have price=0, paid template must have price>0")
)

// TemplateService 模板服务
type TemplateService struct {
	templateRepo *repository.TemplateRepo
	presigner    *oss.Presigner
	idGen        *idgen.Generator
}

// NewTemplateService 创建模板服务实例。
func NewTemplateService(templateRepo *repository.TemplateRepo, presigner *oss.Presigner, idGen *idgen.Generator) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		presigner:    presigner,
		idGen:        idGen,
	}
}

// TemplateWithURL 模板 + 缩略图预签名 URL
type TemplateWithURL struct {
	Template             model.TosTemplate
	ThumbnailDownloadURL string
}

// CreateTemplate 创建模板：校验 is_free=1 时 price=0、is_free=0 时 price>0。
func (s *TemplateService) CreateTemplate(template *model.TosTemplate) error {
	if template.IsFree == 1 && template.Price != 0 {
		return ErrPriceInvalid
	}
	if template.IsFree == 0 && template.Price <= 0 {
		return ErrPriceInvalid
	}

	now := time.Now()
	template.TemplateNo = s.idGen.GenerateTemplateNo()
	template.CreatedAt = now
	template.UpdatedAt = now

	return s.templateRepo.Create(template)
}

// UpdateTemplate 更新模板基本信息。
func (s *TemplateService) UpdateTemplate(template *model.TosTemplate) error {
	template.UpdatedAt = time.Now()
	return s.templateRepo.Update(template)
}

// UpdateStatus 上下架模板。
func (s *TemplateService) UpdateStatus(id uint64, status int8) error {
	return s.templateRepo.UpdateStatus(id, status)
}

// GetTemplate 获取模板详情，含缩略图预签名 URL。
func (s *TemplateService) GetTemplate(id uint64) (*TemplateWithURL, error) {
	template, err := s.templateRepo.FindByID(id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	url, _ := s.presigner.PresignThumbnail(template.ThumbnailOssKey)

	return &TemplateWithURL{
		Template:             *template,
		ThumbnailDownloadURL: url,
	}, nil
}

// ListTemplates 分页查询模板，含缩略图预签名 URL。
func (s *TemplateService) ListTemplates(status *int8, offset, limit int) ([]TemplateWithURL, int64, error) {
	templates, total, err := s.templateRepo.ListTemplates(status, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]TemplateWithURL, len(templates))
	for i, t := range templates {
		url, _ := s.presigner.PresignThumbnail(t.ThumbnailOssKey)
		result[i] = TemplateWithURL{
			Template:             t,
			ThumbnailDownloadURL: url,
		}
	}

	return result, total, nil
}