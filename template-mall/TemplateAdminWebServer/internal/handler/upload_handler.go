package handler

import (
	"net/http"
	"strings"
	"time"

	"template-mall/TemplateAdminWebServer/internal/oss"

	"github.com/gin-gonic/gin"
)

// UploadHandler 上传凭证签发 + 确认 handler。
type UploadHandler struct {
	sts        *oss.STSIssuer
	headObject *oss.HeadObject
	objectKey  *oss.ObjectKey
}

// NewUploadHandler 创建上传 handler。
func NewUploadHandler(sts *oss.STSIssuer, head *oss.HeadObject, objKey *oss.ObjectKey) *UploadHandler {
	return &UploadHandler{sts: sts, headObject: head, objectKey: objKey}
}

const (
	MaxTemplateSize = 5 * 1024 * 1024 // 5MB
	MaxThumbSize    = 1 * 1024 * 1024 // 1MB
)

var allowedTemplateTypes = map[string]bool{
	"ppt":  true,
	"pptx": true,
	"doc":  true,
	"docx": true,
}

var allowedThumbTypes = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
}

// CredentialRequest 签发 STS 凭证请求。
type CredentialRequest struct {
	FileType string `json:"file_type" binding:"required"` // ppt/pptx/doc/docx 或 jpg/png
	FileSize int64  `json:"file_size" binding:"required"` // 字节
	// thumbnail 模式：file_type=jpg/png，size ≤ 1MB
	IsThumbnail bool `json:"is_thumbnail"`
}

// CredentialResponse 签发 STS 凭证响应。
type CredentialResponse struct {
	ObjectKey       string    `json:"object_key"`
	AccessKeyID     string    `json:"access_key_id"`
	AccessKeySecret string    `json:"access_key_secret"`
	SecurityToken   string    `json:"security_token"`
	Expiration      time.Time `json:"expiration"`
	Bucket          string    `json:"bucket"`
	Endpoint        string    `json:"endpoint"`
	Region          string    `json:"region"`
}

// Credential POST /api/admin/upload/credential
// 校验 file_type 与 size → 生成 Object Key → 签发 STS（5min）。
func (h *UploadHandler) Credential(c *gin.Context) {
	var req CredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	fileType := strings.ToLower(strings.TrimPrefix(req.FileType, "."))
	var prefix string
	var maxSize int64

	if req.IsThumbnail {
		if !allowedThumbTypes[fileType] {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "thumbnail must be jpg/png"})
			return
		}
		if req.FileSize > MaxThumbSize {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "thumbnail size exceeds 1MB"})
			return
		}
		prefix = "thumbnails"
		maxSize = MaxThumbSize
	} else {
		if !allowedTemplateTypes[fileType] {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "template must be ppt/pptx/doc/docx"})
			return
		}
		if req.FileSize > MaxTemplateSize {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "template size exceeds 5MB"})
			return
		}
		prefix = "templates"
		maxSize = MaxTemplateSize
	}

	// 生成 Object Key
	objectKey := h.objectKey.Generate(prefix, fileType)

	// 签发 STS（按 objectKey 限定 PutObject/HeadObject 权限）
	creds, err := h.sts.Issue(objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "issue STS failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": CredentialResponse{
			ObjectKey:       objectKey,
			AccessKeyID:     creds.AccessKeyID,
			AccessKeySecret: creds.AccessKeySecret,
			SecurityToken:   creds.SecurityToken,
			Expiration:      creds.Expiration,
			Bucket:          creds.Bucket,
			Endpoint:        creds.Endpoint,
			Region:          creds.Region,
		},
		"max_size": maxSize, // 提示前端最大允许大小
	})
}

// ConfirmRequest 确认上传请求。
type ConfirmRequest struct {
	ObjectKey string `json:"object_key" binding:"required"`
	FileSize  int64  `json:"file_size"`
}

// Confirm POST /api/admin/upload/confirm
// HeadObject 确认存在 + Key 与签发一致 + size ≤ 5MB。
func (h *UploadHandler) Confirm(c *gin.Context) {
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// 校验 Object Key 前缀
	prefix, _, _, _ := oss.ParseObjectKey(req.ObjectKey)
	if prefix != "templates" && prefix != "thumbnails" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid object key prefix"})
		return
	}

	// 决定最大 size
	var maxSize int64 = MaxTemplateSize
	if prefix == "thumbnails" {
		maxSize = MaxThumbSize
	}

	// HeadObject 确认
	exists, size, err := h.headObject.Confirm(req.ObjectKey, maxSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "head object failed: " + err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "object not found in OSS"})
		return
	}
	if size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "object size exceeds limit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"object_key":      req.ObjectKey,
			"file_size":       size,
			"confirmed":       true,
		},
	})
}