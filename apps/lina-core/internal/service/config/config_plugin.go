package config

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// PluginConfig holds plugin-related host configuration.
type PluginConfig struct {
	Dynamic PluginDynamicConfig `json:"dynamic"` // Dynamic contains dynamic plugin storage settings.
	Runtime PluginDynamicConfig `json:"runtime"` // Runtime keeps legacy config compatibility for older runtime keys.
}

// PluginDynamicConfig holds dynamic plugin storage configuration.
type PluginDynamicConfig struct {
	StoragePath string `json:"storagePath"` // StoragePath is the directory used to discover and store dynamic wasm packages.
}

// GetPlugin reads plugin config from configuration file.
func (s *Service) GetPlugin(ctx context.Context) *PluginConfig {
	cfg := &PluginConfig{
		Dynamic: PluginDynamicConfig{
			StoragePath: "temp/runtime",
		},
	}
	_ = g.Cfg().MustGet(ctx, "plugin").Scan(cfg)

	cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Dynamic.StoragePath)
	if cfg.Dynamic.StoragePath == "" {
		cfg.Dynamic.StoragePath = strings.TrimSpace(cfg.Runtime.StoragePath)
	}
	if cfg.Dynamic.StoragePath == "" {
		cfg.Dynamic.StoragePath = "temp/runtime"
	}
	return cfg
}

// GetPluginDynamicStoragePath returns the normalized dynamic wasm storage directory.
func (s *Service) GetPluginDynamicStoragePath(ctx context.Context) string {
	return filepath.Clean(s.GetPlugin(ctx).Dynamic.StoragePath)
}
