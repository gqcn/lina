package config

import (
	"context"
)

// InitConfig holds database initialization configuration.
type InitConfig struct {
	SqlDir string `json:"sqlDir"` // SQL file directory
}

// GetInit reads initialization config from configuration file.
func (s *Service) GetInit(ctx context.Context) *InitConfig {
	cfg := &InitConfig{
		SqlDir: "manifest/sql",
	}
	mustScanConfig(ctx, "init", cfg)
	return cfg
}
