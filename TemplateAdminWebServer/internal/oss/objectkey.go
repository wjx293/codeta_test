package oss

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ObjectKey 生成 OSS Object Key。
// 规则：{prefix}/{yyyyMM}/{uuid}.{ext}
// {yyyyMM} 按 Asia/Shanghai 时区计算（宪法 CL-6）。
type ObjectKey struct{}

// NewObjectKey 创建生成器。
func NewObjectKey() *ObjectKey {
	return &ObjectKey{}
}

// Generate 生成 Object Key。
// prefix: "templates" 或 "thumbnails"
// ext: 文件扩展名（不含点），如 "pptx"、"jpg"
func (o *ObjectKey) Generate(prefix, ext string) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.UTC
	}
	yyyyMM := time.Now().In(loc).Format("200601")
	uuidStr := uuid.New().String()
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	return fmt.Sprintf("%s/%s/%s.%s", prefix, yyyyMM, uuidStr, ext)
}

// ValidatePrefix 校验 prefix 是否合法。
func (o *ObjectKey) ValidatePrefix(prefix string) bool {
	return prefix == "templates" || prefix == "thumbnails"
}

// ParseObjectKey 解析 Object Key，返回 (prefix, yyyyMM, uuid, ext)。
func ParseObjectKey(key string) (prefix, yyyyMM, uuidStr, ext string) {
	base := filepath.Base(key)
	ext = strings.TrimPrefix(filepath.Ext(base), ".")
	name := strings.TrimSuffix(base, filepath.Ext(base))
	parts := strings.SplitN(key, "/", 3)
	if len(parts) >= 1 {
		prefix = parts[0]
	}
	if len(parts) >= 2 {
		yyyyMM = parts[1]
	}
	uuidStr = name
	return
}