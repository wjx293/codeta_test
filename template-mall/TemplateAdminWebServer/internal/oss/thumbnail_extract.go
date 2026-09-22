package oss

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
)

// 与 upload_handler 中限制保持一致。
const (
	MaxTemplateObjectSize = 5 * 1024 * 1024
	MaxThumbObjectSize    = 1 * 1024 * 1024
)

// 常见 Office OOXML 内嵌缩略图路径（pptx/docx 等为 zip 包）。
var embeddedThumbnailPaths = []string{
	"docProps/thumbnail.jpeg",
	"docProps/thumbnail.jpg",
	"docProps/thumbnail.png",
	"ppt/thumbnail.jpeg",
	"ppt/thumbnail.jpg",
	"ppt/thumbnail.png",
}

// ExtractEmbeddedThumbnail 从 pptx/docx 等 OOXML 文件中提取内嵌缩略图。
// 返回图片数据、扩展名（jpg/png）、错误。
// 旧版二进制 ppt/doc 无内嵌缩略图时返回 ("", "", nil)。
func ExtractEmbeddedThumbnail(fileData []byte, fileType string) ([]byte, string, error) {
	ft := strings.ToLower(strings.TrimPrefix(fileType, "."))
	switch ft {
	case "pptx", "docx":
		return extractFromZip(fileData)
	case "ppt", "doc":
		// 旧版 OLE 格式需 LibreOffice 等外部工具，当前不支持自动提取
		return nil, "", nil
	default:
		return nil, "", fmt.Errorf("unsupported file type for thumbnail extraction: %s", fileType)
	}
}

func extractFromZip(fileData []byte) ([]byte, string, error) {
	reader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return nil, "", fmt.Errorf("open zip: %w", err)
	}

	index := make(map[string]*zip.File, len(reader.File))
	for _, f := range reader.File {
		index[path.Clean(f.Name)] = f
	}

	for _, candidate := range embeddedThumbnailPaths {
		f, ok := index[candidate]
		if !ok {
			continue
		}
		data, err := readZipEntry(f, MaxThumbObjectSize+1)
		if err != nil {
			continue
		}
		if int64(len(data)) > MaxThumbObjectSize {
			continue
		}
		ext := strings.TrimPrefix(path.Ext(candidate), ".")
		if ext == "jpeg" {
			ext = "jpg"
		}
		return data, ext, nil
	}

	// pptx：尝试取 ppt/media 下第一张图作为封面
	if img, ext, ok := firstPptMediaImage(reader); ok {
		return img, ext, nil
	}

	return nil, "", nil
}

func readZipEntry(f *zip.File, maxSize int64) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, maxSize))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func firstPptMediaImage(reader *zip.Reader) ([]byte, string, bool) {
	for _, f := range reader.File {
		if !strings.HasPrefix(f.Name, "ppt/media/") {
			continue
		}
		ext := strings.ToLower(strings.TrimPrefix(path.Ext(f.Name), "."))
		if ext != "png" && ext != "jpg" && ext != "jpeg" {
			continue
		}
		data, err := readZipEntry(f, MaxThumbObjectSize+1)
		if err != nil || int64(len(data)) > MaxThumbObjectSize || len(data) == 0 {
			continue
		}
		if ext == "jpeg" {
			ext = "jpg"
		}
		return data, ext, true
	}
	return nil, "", false
}
