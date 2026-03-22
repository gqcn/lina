package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// MonitorConfig holds server monitoring configuration.
type MonitorConfig struct {
	IntervalSeconds  int `json:"intervalSeconds"`
	RetentionMinutes int `json:"retentionMinutes"`
}

// GetMonitor reads monitor config from configuration file.
func (s *Service) GetMonitor(ctx context.Context) *MonitorConfig {
	cfg := &MonitorConfig{
		IntervalSeconds:  30,
		RetentionMinutes: 60,
	}
	_ = g.Cfg().MustGet(ctx, "monitor").Scan(cfg)
	return cfg
}
