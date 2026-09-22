package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"template-mall/TemplateAdminWebServer/internal/client"
	"template-mall/TemplateAdminWebServer/internal/oss"

	pb "template-mall/TemplateOrderServer/api/proto"

	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板管理 handler。
type TemplateHandler struct {
	grpcClient  *client.GRPCClient
	headObject  *oss.HeadObject
	thumbGen    *oss.ThumbnailGenerator
}

// NewTemplateHandler 创建模板 handler。
func NewTemplateHandler(grpcClient *client.GRPCClient, head *oss.HeadObject, thumbGen *oss.ThumbnailGenerator) *TemplateHandler {
	return &TemplateHandler{grpcClient: grpcClient, headObject: head, thumbGen: thumbGen}
}

// CreateTemplateRequest 创建模板请求。
type CreateTemplateRequest struct {
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	IsFree            int32  `json:"is_free" binding:"oneof=0 1"` // 0=付费 1=免费（不能用 required，0 会被当成空值）
	Price             int64  `json:"price"`                          // 分
	FileType          string `json:"file_type" binding:"required"`
	FileSize          int64  `json:"file_size" binding:"required"`
	FileOssKey        string `json:"file_oss_key" binding:"required"`
	ThumbnailOssKey   string `json:"thumbnail_oss_key"`
	ThumbnailFileSize int64  `json:"thumbnail_file_size"`
}

// Create POST /api/admin/templates
func (h *TemplateHandler) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	thumbKey := req.ThumbnailOssKey
	thumbSize := req.ThumbnailFileSize
	if thumbKey == "" && h.thumbGen != nil {
		if key, size, err := h.thumbGen.FromTemplateFile(req.FileOssKey, req.FileType); err != nil {
			// 缩略图提取失败不阻断创建，仅记录日志
			log.Printf("auto thumbnail skipped for %s: %v", req.FileOssKey, err)
		} else if key != "" {
			thumbKey = key
			thumbSize = size
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	resp, err := h.grpcClient.Stub.CreateTemplate(ctx, &pb.CreateTemplateRequest{
		Name:              req.Name,
		Description:       req.Description,
		IsFree:            req.IsFree == 1,
		Price:             req.Price,
		OssType:           "aliyun-oss",
		BucketName:        "",
		OriginalFilename:  req.Name,
		FileType:          req.FileType,
		FileSize:          req.FileSize,
		FileOssKey:        req.FileOssKey,
		ThumbnailOssKey:   thumbKey,
		ThumbnailFileSize: thumbSize,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "template created"})
}

// UpdateRequest 修改模板请求。
type UpdateRequest struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsFree            int32  `json:"is_free"`
	Price             int64  `json:"price"`
	ThumbnailOssKey   string `json:"thumbnail_oss_key"`
	ThumbnailFileSize int64  `json:"thumbnail_file_size"`
}

// Update PUT /api/admin/templates/:id
func (h *TemplateHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.grpcClient.Stub.UpdateTemplate(ctx, &pb.UpdateTemplateRequest{
		Id:                id,
		Name:              req.Name,
		Description:       req.Description,
		IsFree:            req.IsFree == 1,
		Price:             req.Price,
		ThumbnailOssKey:   req.ThumbnailOssKey,
		ThumbnailFileSize: req.ThumbnailFileSize,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "template updated"})
}

// UpdateStatusRequest 上下架请求。
type UpdateStatusRequest struct {
	Status int32 `json:"status" binding:"oneof=0 1"` // 0=下架 1=上架
}

// UpdateStatus PUT /api/admin/templates/:id/status
func (h *TemplateHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if req.Status != 0 && req.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "status must be 0 or 1"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.grpcClient.Stub.UpdateTemplateStatus(ctx, &pb.UpdateTemplateStatusRequest{
		Id:     id,
		Status: templateStatusFromJSON(req.Status),
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Code, "message": resp.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "status updated"})
}

// List GET /api/admin/templates
// 查询参数：status=-1(全部)|0(下架)|1(上架)，page，page_size
func (h *TemplateHandler) List(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	pageVal := int32(page)
	pageSizeVal := int32(pageSize)
	req := &pb.ListTemplatesRequest{
		Page:     &pageVal,
		PageSize: &pageSizeVal,
	}
	if status >= 0 {
		req.Status = templateStatusFromQuery(status)
	}
	resp, err := h.grpcClient.Stub.ListTemplates(ctx, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
		return
	}
	if resp.Base != nil && resp.Base.Code != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.Base.Code, "message": resp.Base.Message})
		return
	}

	templates := make([]gin.H, 0, len(resp.Templates))
	for _, t := range resp.Templates {
		templates = append(templates, templateToJSON(t))
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"templates": templates, "total": resp.Total},
	})
}

// templateToJSON 将 proto Template 转换为 JSON 响应。
func templateToJSON(t *pb.Template) gin.H {
	var createdAt, updatedAt string
	if t.CreatedAt != nil {
		createdAt = t.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if t.UpdatedAt != nil {
		updatedAt = t.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	return gin.H{
		"id":                   t.Id,
		"template_no":          t.TemplateNo,
		"name":                 t.Name,
		"description":          t.Description,
		"is_free":              isFreeJSON(t.IsFree),
		"price":                t.Price,
		"status":               templateStatusJSON(t.Status),
		"file_type":            t.FileType,
		"file_size":            t.FileSize,
		"thumbnail_download_url": t.ThumbnailDownloadUrl,
		"created_at":           createdAt,
		"updated_at":           updatedAt,
	}
}