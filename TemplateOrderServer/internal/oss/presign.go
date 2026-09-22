// Package oss 提供对象存储预签名 URL 生成功能。
package oss

import (
	"fmt"
	"time"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Presigner 预签名 URL 生成器。
type Presigner struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
}

// NewPresigner 创建预签名生成器。
func NewPresigner(endpoint, accessKey, secretKey, bucket string) *Presigner {
	return &Presigner{
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
	}
}

// PresignDownload 生成下载预签名 URL（5min TTL）。
// ossKey 为空时返回空串不报错。
func (p *Presigner) PresignDownload(ossKey string, ttl time.Duration) (string, error) {
	if ossKey == "" {
		return "", nil
	}
	if !p.configured() {
		return fmt.Sprintf("https://%s.%s/%s?presigned=placeholder", p.bucket, p.endpoint, ossKey), nil
	}

	bucket, err := p.bucketClient()
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return bucket.SignURL(ossKey, aliyunoss.HTTPGet, int64(ttl.Seconds()))
}

// PresignThumbnail 生成缩略图预签名 URL（5min 默认 TTL）。
// ossKey 为空时返回空串不报错。
func (p *Presigner) PresignThumbnail(ossKey string) (string, error) {
	return p.PresignDownload(ossKey, 5*time.Minute)
}
