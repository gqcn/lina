package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ElectionConfig holds leader election configuration.
type ElectionConfig struct {
	Lease         time.Duration `json:"lease"`         // Lock lease duration
	RenewInterval time.Duration `json:"renewInterval"` // Lease renewal interval
}

// GetElection reads election config from configuration file.
func (s *Service) GetElection(ctx context.Context) *ElectionConfig {
	cfg := &ElectionConfig{
		Lease:         30 * time.Second,
		RenewInterval: 10 * time.Second,
	}
	_ = g.Cfg().MustGet(ctx, "election").Scan(cfg)
	return cfg
}
