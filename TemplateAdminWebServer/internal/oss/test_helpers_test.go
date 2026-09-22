package oss

import "template-mall/TemplateAdminWebServer/internal/config"

// cfgForTest 构造测试用 config（无 OSS 配置 → mock 模式）。
func cfgForTest() *config.Config {
	return &config.Config{
		OSSEndpoint:        "",
		OSSAccessKeyID:     "",
		OSSAccessKeySecret: "",
		OSSBucket:          "test-bucket",
		OSSRegion:          "cn-beijing",
	}
}