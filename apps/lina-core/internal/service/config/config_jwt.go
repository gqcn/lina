package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// JwtConfig holds JWT authentication configuration.
type JwtConfig struct {
	Secret     string `json:"secret"`     // JWT secret key
	ExpireHour int    `json:"expireHour"` // Expiration time (hours)
}

// GetJwt reads JWT config from configuration file.
func (s *Service) GetJwt(ctx context.Context) *JwtConfig {
	cfg := &JwtConfig{}
	_ = g.Cfg().MustGet(ctx, "jwt").Scan(cfg)
	return cfg
}
