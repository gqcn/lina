package config

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginConfig holds plugin-related host configuration.
type PluginConfig struct {
	Runtime PluginRuntimeConfig `json:"runtime"` // Runtime contains runtime wasm plugin storage settings.
}

// PluginRuntimeConfig holds runtime wasm plugin storage configuration.
type PluginRuntimeConfig struct {
	StoragePath string `json:"storagePath"` // StoragePath is the directory used to discover and store runtime wasm packages.
}

// GetPlugin reads plugin config from configuration file.
func (s *Service) GetPlugin(ctx context.Context) *PluginConfig {
	cfg := &PluginConfig{
		Runtime: PluginRuntimeConfig{
			StoragePath: "temp/runtime",
		},
	}
	_ = g.Cfg().MustGet(ctx, "plugin").Scan(cfg)
	cfg.Runtime.StoragePath = strings.TrimSpace(cfg.Runtime.StoragePath)
	if cfg.Runtime.StoragePath == "" {
		cfg.Runtime.StoragePath = "temp/runtime"
	}
	return cfg
}

// GetPluginRuntimeStoragePath returns the normalized runtime wasm storage directory.
func (s *Service) GetPluginRuntimeStoragePath(ctx context.Context) string {
	return filepath.Clean(s.GetPlugin(ctx).Runtime.StoragePath)
}
