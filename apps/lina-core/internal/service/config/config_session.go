package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// SessionConfig holds session management configuration.
type SessionConfig struct {
	TimeoutHour   int `json:"timeoutHour"`   // Session timeout (hours)
	CleanupMinute int `json:"cleanupMinute"` // Cleanup interval (minutes)
}

// GetSession reads session config from configuration file.
func (s *Service) GetSession(ctx context.Context) *SessionConfig {
	cfg := &SessionConfig{
		TimeoutHour:   24,
		CleanupMinute: 5,
	}
	_ = g.Cfg().MustGet(ctx, "session").Scan(cfg)
	return cfg
}
