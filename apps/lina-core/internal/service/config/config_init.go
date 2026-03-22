package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// InitConfig holds database initialization configuration.
type InitConfig struct {
	SqlDir string `json:"sqlDir"`
}

// GetInit reads initialization config from configuration file.
func (s *Service) GetInit(ctx context.Context) *InitConfig {
	cfg := &InitConfig{
		SqlDir: "manifest/sql",
	}
	_ = g.Cfg().MustGet(ctx, "init").Scan(cfg)
	return cfg
}
