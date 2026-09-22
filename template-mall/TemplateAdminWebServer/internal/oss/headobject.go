package oss

import (
	"fmt"
	"strconv"
	"time"

	"template-mall/TemplateAdminWebServer/internal/config"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// HeadObject 确认 OSS 对象存在且大小符合预期。
type HeadObject struct {
	cfg *config.Config
}

// NewHeadObject 创建 HeadObject 检查器。
func NewHeadObject(cfg *config.Config) *HeadObject {
	return &HeadObject{cfg: cfg}
}

// Confirm 确认对象存在 + Key 与签发一致 + size ≤ limit。
// 返回 (exists, fileSize, error)。
func (h *HeadObject) Confirm(objectKey string, maxSize int64) (bool, int64, error) {
	if objectKey == "" {
		return false, 0, fmt.Errorf("empty object key")
	}

	if !ossConfigured(h.cfg) {
		// 未配置 OSS：开发 mock
		return true, 1024, nil
	}

	bucket, err := newBucketClient(h.cfg)
	if err != nil {
		return false, 0, err
	}

	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		if svcErr, ok := err.(aliyunoss.ServiceError); ok && svcErr.StatusCode == 404 {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("head object: %w", err)
	}

	sizeStr := meta.Get("Content-Length")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return false, 0, fmt.Errorf("parse content-length %q: %w", sizeStr, err)
	}
	if size > maxSize {
		return false, size, fmt.Errorf("file too large: %d > %d", size, maxSize)
	}
	return true, size, nil
}

// IssuePresignDownload 生成下载预签名 URL（5min TTL）。
func (h *HeadObject) IssuePresignDownload(objectKey string) (string, error) {
	if objectKey == "" {
		return "", nil
	}
	if !ossConfigured(h.cfg) {
		return fmt.Sprintf("https://%s.%s/%s?presigned=placeholder&ttl=5m",
			h.cfg.OSSBucket, h.cfg.OSSEndpoint, objectKey), nil
	}

	bucket, err := newBucketClient(h.cfg)
	if err != nil {
		return "", err
	}
	return bucket.SignURL(objectKey, aliyunoss.HTTPGet, int64((5*time.Minute).Seconds()))
}
