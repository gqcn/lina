package config

import (
	"context"
)

// OpenApiConfig holds OpenAPI documentation configuration.
type OpenApiConfig struct {
	Title             string `json:"title"`             // API title
	Description       string `json:"description"`       // API description
	Version           string `json:"version"`           // API version
	ServerUrl         string `json:"serverUrl"`         // Server URL
	ServerDescription string `json:"serverDescription"` // Server description
}

// GetOpenApi reads OpenAPI config from configuration file.
func (s *Service) GetOpenApi(ctx context.Context) *OpenApiConfig {
	cfg := &OpenApiConfig{
		Title:             "Lina Admin API",
		Version:           "v1.0.0",
		ServerDescription: "API Server",
	}
	mustScanConfig(ctx, "openapi", cfg)
	return cfg
}
