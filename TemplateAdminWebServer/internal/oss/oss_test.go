package oss

import (
	"strings"
	"testing"
	"time"
)

func TestObjectKey_Generate(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	expectedMonth := time.Now().In(loc).Format("200601")

	gen := NewObjectKey()
	key := gen.Generate("templates", "pptx")

	if !strings.HasPrefix(key, "templates/"+expectedMonth+"/") {
		t.Fatalf("expected prefix templates/%s/, got %s", expectedMonth, key)
	}
	if !strings.HasSuffix(key, ".pptx") {
		t.Fatalf("expected suffix .pptx, got %s", key)
	}
	// UUID 部分长度应为 36（uuid v4 含横线）
	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %s", len(parts), key)
	}
	if len(parts[2]) != len("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.pptx") {
		t.Fatalf("unexpected uuid.ext length: %s", parts[2])
	}
}

func TestObjectKey_GenerateThumbnail(t *testing.T) {
	gen := NewObjectKey()
	key := gen.Generate("thumbnails", "jpg")
	if !strings.HasPrefix(key, "thumbnails/") {
		t.Fatalf("expected prefix thumbnails/, got %s", key)
	}
	if !strings.HasSuffix(key, ".jpg") {
		t.Fatalf("expected suffix .jpg, got %s", key)
	}
}

func TestObjectKey_GenerateStripsExtDot(t *testing.T) {
	gen := NewObjectKey()
	key1 := gen.Generate("templates", "docx")
	key2 := gen.Generate("templates", ".docx")
	if !strings.HasSuffix(key1, ".docx") || !strings.HasSuffix(key2, ".docx") {
		t.Fatalf("both should end with .docx: %s, %s", key1, key2)
	}
}

func TestObjectKey_ValidatePrefix(t *testing.T) {
	gen := NewObjectKey()
	if !gen.ValidatePrefix("templates") {
		t.Error("templates should be valid")
	}
	if !gen.ValidatePrefix("thumbnails") {
		t.Error("thumbnails should be valid")
	}
	if gen.ValidatePrefix("uploads") {
		t.Error("uploads should be invalid")
	}
}

func TestParseObjectKey(t *testing.T) {
	key := "templates/202608/abc-123.pptx"
	prefix, yyyyMM, uuidStr, ext := ParseObjectKey(key)
	if prefix != "templates" {
		t.Fatalf("expected prefix templates, got %s", prefix)
	}
	if yyyyMM != "202608" {
		t.Fatalf("expected yyyyMM 202608, got %s", yyyyMM)
	}
	if uuidStr != "abc-123" {
		t.Fatalf("expected uuid abc-123, got %s", uuidStr)
	}
	if ext != "pptx" {
		t.Fatalf("expected ext pptx, got %s", ext)
	}
}

func TestSTSIssuer_Mock(t *testing.T) {
	// 无 OSS 配置：返回 mock 凭证
	issuer := NewSTSIssuer(cfgForTest())
	creds, err := issuer.Issue("templates/202608/test-uuid.pptx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessKeyID == "" {
		t.Error("expected non-empty AccessKeyID")
	}
	if creds.AccessKeySecret == "" {
		t.Error("expected non-empty AccessKeySecret")
	}
	if creds.SecurityToken == "" {
		t.Error("expected non-empty SecurityToken")
	}
	// Expiration 应至少 4 分钟后
	if time.Until(creds.Expiration) < 4*time.Minute {
		t.Errorf("expected expiration >= 4min, got %v", time.Until(creds.Expiration))
	}
}

func TestHeadObject_ConfirmMock(t *testing.T) {
	head := NewHeadObject(cfgForTest())
	exists, size, err := head.Confirm("templates/202608/abc.pptx", 5*1024*1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true in mock mode")
	}
	if size <= 0 {
		t.Error("expected size > 0 in mock mode")
	}
}

func TestHeadObject_ConfirmEmptyKey(t *testing.T) {
	head := NewHeadObject(cfgForTest())
	_, _, err := head.Confirm("", 5*1024*1024)
	if err == nil {
		t.Error("expected error for empty key")
	}
}