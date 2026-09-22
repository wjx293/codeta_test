package oss

import (
	"fmt"

	"template-mall/TemplateAdminWebServer/internal/config"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func ossConfigured(cfg *config.Config) bool {
	return cfg.OSSAccessKeyID != "" &&
		cfg.OSSAccessKeySecret != "" &&
		cfg.OSSEndpoint != "" &&
		cfg.OSSBucket != ""
}

func newBucketClient(cfg *config.Config) (*aliyunoss.Bucket, error) {
	if !ossConfigured(cfg) {
		return nil, fmt.Errorf("oss not configured")
	}
	client, err := aliyunoss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create oss client: %w", err)
	}
	bucket, err := client.Bucket(cfg.OSSBucket)
	if err != nil {
		return nil, fmt.Errorf("open oss bucket: %w", err)
	}
	return bucket, nil
}
