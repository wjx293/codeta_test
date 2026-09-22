package oss

import (
	"fmt"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (p *Presigner) configured() bool {
	return p.accessKey != "" && p.secretKey != "" && p.endpoint != "" && p.bucket != ""
}

func (p *Presigner) bucketClient() (*aliyunoss.Bucket, error) {
	if !p.configured() {
		return nil, fmt.Errorf("oss not configured")
	}
	client, err := aliyunoss.New(p.endpoint, p.accessKey, p.secretKey)
	if err != nil {
		return nil, fmt.Errorf("create oss client: %w", err)
	}
	bucket, err := client.Bucket(p.bucket)
	if err != nil {
		return nil, fmt.Errorf("open oss bucket: %w", err)
	}
	return bucket, nil
}
