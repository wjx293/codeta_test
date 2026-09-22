package oss

import (
	"fmt"
	"strings"
)

// ThumbnailGenerator 从模板文件自动生成缩略图并上传到 OSS。
type ThumbnailGenerator struct {
	head      *HeadObject
	objectKey *ObjectKey
}

// NewThumbnailGenerator 创建缩略图生成器。
func NewThumbnailGenerator(head *HeadObject, objectKey *ObjectKey) *ThumbnailGenerator {
	return &ThumbnailGenerator{head: head, objectKey: objectKey}
}

// FromTemplateFile 若模板文件内含缩略图则提取并上传，返回 OSS key 与大小。
// 无法提取时返回空字符串（非错误）。
func (g *ThumbnailGenerator) FromTemplateFile(fileOssKey, fileType string) (string, int64, error) {
	if fileOssKey == "" {
		return "", 0, nil
	}

	fileData, err := g.head.GetObject(fileOssKey, MaxTemplateObjectSize)
	if err != nil {
		return "", 0, fmt.Errorf("download template for thumbnail: %w", err)
	}

	imgData, ext, err := ExtractEmbeddedThumbnail(fileData, fileType)
	if err != nil {
		return "", 0, err
	}
	if len(imgData) == 0 || ext == "" {
		return "", 0, nil
	}

	thumbKey := g.objectKey.Generate("thumbnails", ext)
	contentType := "image/jpeg"
	if ext == "png" {
		contentType = "image/png"
	}
	if err := g.head.PutObject(thumbKey, imgData, contentType); err != nil {
		return "", 0, fmt.Errorf("upload auto thumbnail: %w", err)
	}

	return thumbKey, int64(len(imgData)), nil
}

// ThumbContentType 根据扩展名返回 Content-Type。
func ThumbContentType(ext string) string {
	if strings.EqualFold(ext, "png") {
		return "image/png"
	}
	return "image/jpeg"
}
