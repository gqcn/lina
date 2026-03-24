package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// JwtConfig holds JWT authentication configuration.
type JwtConfig struct {
	Secret     string `json:"secret"`     // JWT密钥
	ExpireHour int    `json:"expireHour"` // 过期时间（小时）
}

// GetJwt reads JWT config from configuration file.
func (s *Service) GetJwt(ctx context.Context) *JwtConfig {
	cfg := &JwtConfig{}
	_ = g.Cfg().MustGet(ctx, "jwt").Scan(cfg)
	return cfg
}
