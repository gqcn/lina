// This file defines server-monitor configuration loading and duration migration.

package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// MonitorConfig holds server monitoring configuration.
type MonitorConfig struct {
	Interval            time.Duration `json:"interval"`            // Interval is the metrics collection period.
	RetentionMultiplier int           `json:"retentionMultiplier"` // Retention multiplier for stale records.
}

// GetMonitor reads monitor config from configuration file.
func (s *Service) GetMonitor(ctx context.Context) *MonitorConfig {
	cfg := &MonitorConfig{
		Interval:            30 * time.Second,
		RetentionMultiplier: 5,
	}
	if monitorVar := g.Cfg().MustGet(ctx, "monitor"); monitorVar != nil {
		_ = monitorVar.Scan(cfg)
	}
	cfg.Interval = mustLoadDurationConfig(ctx, "monitor.interval", cfg.Interval)
	cfg.Interval = mustValidateSecondAlignedDuration("monitor.interval", cfg.Interval)
	return cfg
}
