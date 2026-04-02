package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// MonitorConfig holds server monitoring configuration.
type MonitorConfig struct {
	IntervalSeconds     int `json:"intervalSeconds"`     // Collection interval (seconds)
	RetentionMultiplier int `json:"retentionMultiplier"` // Retention multiplier for stale records
}

// GetMonitor reads monitor config from configuration file.
func (s *Service) GetMonitor(ctx context.Context) *MonitorConfig {
	cfg := &MonitorConfig{
		IntervalSeconds:     30,
		RetentionMultiplier: 5,
	}
	_ = g.Cfg().MustGet(ctx, "monitor").Scan(cfg)
	return cfg
}
