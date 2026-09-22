package oss

import (
	"bytes"
	"fmt"
	"io"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// GetObject 下载 OSS 对象（不超过 maxSize 字节）。
func (h *HeadObject) GetObject(objectKey string, maxSize int64) ([]byte, error) {
	if objectKey == "" {
		return nil, fmt.Errorf("empty object key")
	}
	if !ossConfigured(h.cfg) {
		return nil, fmt.Errorf("oss not configured")
	}

	bucket, err := newBucketClient(h.cfg)
	if err != nil {
		return nil, err
	}

	body, err := bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer body.Close()

	data, err := io.ReadAll(io.LimitReader(body, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file too large: exceeds %d bytes", maxSize)
	}
	return data, nil
}

// PutObject 上传对象到 OSS。
func (h *HeadObject) PutObject(objectKey string, data []byte, contentType string) error {
	if objectKey == "" {
		return fmt.Errorf("empty object key")
	}
	if !ossConfigured(h.cfg) {
		return fmt.Errorf("oss not configured")
	}

	bucket, err := newBucketClient(h.cfg)
	if err != nil {
		return err
	}

	opts := []aliyunoss.Option{}
	if contentType != "" {
		opts = append(opts, aliyunoss.ContentType(contentType))
	}
	if err := bucket.PutObject(objectKey, bytes.NewReader(data), opts...); err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}
