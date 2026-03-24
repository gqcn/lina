package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// MonitorConfig holds server monitoring configuration.
type MonitorConfig struct {
	IntervalSeconds int `json:"intervalSeconds"` // 采集间隔（秒）
}

// GetMonitor reads monitor config from configuration file.
func (s *Service) GetMonitor(ctx context.Context) *MonitorConfig {
	cfg := &MonitorConfig{
		IntervalSeconds: 30,
	}
	_ = g.Cfg().MustGet(ctx, "monitor").Scan(cfg)
	return cfg
}
