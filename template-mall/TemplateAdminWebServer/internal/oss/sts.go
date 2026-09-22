// Package oss 提供 OSS STS 凭证签发与 HeadObject 确认。
package oss

import (
	"encoding/json"
	"fmt"
	"time"

	"template-mall/TemplateAdminWebServer/internal/config"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
)

// 阿里云 AssumeRole 最短 900 秒（15 分钟），见 STS API DurationSeconds 限制。
const stsDurationSeconds = 900

// STSCredentials STS 临时凭证。
type STSCredentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
	Expiration      time.Time
	Bucket          string
	Endpoint        string
	Region          string
}

// STSIssuer STS 凭证签发器。
type STSIssuer struct {
	cfg *config.Config
}

// NewSTSIssuer 创建 STS 签发器。
func NewSTSIssuer(cfg *config.Config) *STSIssuer {
	return &STSIssuer{cfg: cfg}
}

// Issue 为指定 objectKey 签发 STS 临时上传凭证（有效期 5 分钟）。
// 未配置 OSS 时返回 mock；已配置 OSS 时必须配置 STS_ROLE_ARN，通过 AssumeRole 换取临时三元组。
// 永久 AccessKey Secret 仅留在服务端，不下发前端。
func (s *STSIssuer) Issue(objectKey string) (*STSCredentials, error) {
	if objectKey == "" {
		return nil, fmt.Errorf("object key is required")
	}

	expiration := time.Now().Add(stsDurationSeconds * time.Second)

	if !ossConfigured(s.cfg) {
		return &STSCredentials{
			AccessKeyID:     "mock-access-key-id",
			AccessKeySecret: "mock-access-key-secret",
			SecurityToken:   "mock-security-token",
			Expiration:      expiration,
			Bucket:          s.cfg.OSSBucket,
			Endpoint:        s.cfg.OSSEndpoint,
			Region:          s.cfg.OSSRegion,
		}, nil
	}

	if s.cfg.STSRoleARN == "" {
		return nil, fmt.Errorf("STS_ROLE_ARN is required when OSS is configured")
	}

	return s.assumeRole(objectKey, expiration)
}

func (s *STSIssuer) assumeRole(objectKey string, fallbackExpiration time.Time) (*STSCredentials, error) {
	region := s.cfg.OSSRegion
	if region == "" {
		region = "cn-hangzhou"
	}

	client, err := sts.NewClientWithAccessKey(region, s.cfg.OSSAccessKeyID, s.cfg.OSSAccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create sts client: %w", err)
	}

	policy, err := buildObjectPolicy(s.cfg.OSSBucket, objectKey)
	if err != nil {
		return nil, err
	}

	sessionName := s.cfg.STSSessionName
	if sessionName == "" {
		sessionName = "template-mall-admin"
	}

	req := sts.CreateAssumeRoleRequest()
	req.Scheme = "https"
	req.RoleArn = s.cfg.STSRoleARN
	req.RoleSessionName = sessionName
	req.DurationSeconds = requests.NewInteger(stsDurationSeconds)
	req.Policy = policy

	resp, err := client.AssumeRole(req)
	if err != nil {
		return nil, fmt.Errorf("assume role: %w", err)
	}
	if resp.Credentials.AccessKeyId == "" || resp.Credentials.AccessKeySecret == "" || resp.Credentials.SecurityToken == "" {
		return nil, fmt.Errorf("assume role returned incomplete credentials")
	}

	expiration := fallbackExpiration
	if resp.Credentials.Expiration != "" {
		if t, parseErr := time.Parse(time.RFC3339, resp.Credentials.Expiration); parseErr == nil {
			expiration = t
		}
	}

	return &STSCredentials{
		AccessKeyID:     resp.Credentials.AccessKeyId,
		AccessKeySecret: resp.Credentials.AccessKeySecret,
		SecurityToken:   resp.Credentials.SecurityToken,
		Expiration:      expiration,
		Bucket:          s.cfg.OSSBucket,
		Endpoint:        s.cfg.OSSEndpoint,
		Region:          region,
	}, nil
}

func buildObjectPolicy(bucket, objectKey string) (string, error) {
	resource := fmt.Sprintf("acs:oss:*:*:%s/%s", bucket, objectKey)
	policy := map[string]any{
		"Version": "1",
		"Statement": []map[string]any{
			{
				"Effect":   "Allow",
				"Action":   []string{"oss:PutObject"},
				"Resource": []string{resource},
			},
			{
				"Effect":   "Allow",
				"Action":   []string{"oss:HeadObject"},
				"Resource": []string{resource},
			},
		},
	}
	b, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("marshal sts policy: %w", err)
	}
	return string(b), nil
}
