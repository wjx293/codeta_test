package oss

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestExtractEmbeddedThumbnail_docProps(t *testing.T) {
	payload := []byte("fake-jpeg-data")
	zipData := buildTestZip(t, map[string][]byte{
		"docProps/thumbnail.jpeg": payload,
		"[Content_Types].xml":     []byte("<Types/>"),
	})

	got, ext, err := ExtractEmbeddedThumbnail(zipData, "pptx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != "jpg" {
		t.Fatalf("expected jpg, got %s", ext)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestExtractEmbeddedThumbnail_pptMediaFallback(t *testing.T) {
	payload := []byte("png-bytes")
	zipData := buildTestZip(t, map[string][]byte{
		"ppt/media/image1.png": payload,
	})

	got, ext, err := ExtractEmbeddedThumbnail(zipData, "pptx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != "png" {
		t.Fatalf("expected png, got %s", ext)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestExtractEmbeddedThumbnail_legacyPptSkipped(t *testing.T) {
	got, ext, err := ExtractEmbeddedThumbnail([]byte{0xD0, 0xCF}, "ppt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil || ext != "" {
		t.Fatalf("expected empty result for legacy ppt")
	}
}

func buildTestZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
