package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// OpenApiConfig holds OpenAPI documentation configuration.
type OpenApiConfig struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	Version           string `json:"version"`
	ServerUrl         string `json:"serverUrl"`
	ServerDescription string `json:"serverDescription"`
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
