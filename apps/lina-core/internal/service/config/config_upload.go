package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	Path    string `json:"path"`    // 上传目录
	MaxSize int64  `json:"maxSize"` // 最大文件大小（MB）
}

// GetUpload reads upload config from configuration file.
func (s *Service) GetUpload(ctx context.Context) *UploadConfig {
	cfg := &UploadConfig{
		Path:    "temp/upload",
		MaxSize: 10,
	}
	_ = g.Cfg().MustGet(ctx, "upload").Scan(cfg)
	return cfg
}
