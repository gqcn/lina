package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// OpenApiConfig holds OpenAPI documentation configuration.
type OpenApiConfig struct {
	Title             string `json:"title"`             // API标题
	Description       string `json:"description"`       // API描述
	Version           string `json:"version"`           // API版本
	ServerUrl         string `json:"serverUrl"`         // 服务器URL
	ServerDescription string `json:"serverDescription"` // 服务器描述
}

// GetOpenApi reads OpenAPI config from configuration file.
func (s *Service) GetOpenApi(ctx context.Context) *OpenApiConfig {
	cfg := &OpenApiConfig{
		Title:             "Lina Admin API",
		Version:           "v1.0.0",
		ServerDescription: "API Server",
	}
	_ = g.Cfg().MustGet(ctx, "openapi").Scan(cfg)
	return cfg
}
