package cron

import (
	"context"

	"lina-core/internal/service/config"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"
)

// Service manages all scheduled/cron tasks.
type Service struct {
	configSvc    *config.Service    // 配置服务
	serverMonSvc *servermon.Service // 服务器监控服务
	sessionStore session.Store      // 会话存储
}

// New creates and returns a new Service instance.
func New(sessionStore session.Store) *Service {
	return &Service{
		configSvc:    config.New(),
		serverMonSvc: servermon.New(),
		sessionStore: sessionStore,
	}
}

// Start registers and starts all cron jobs.
func (s *Service) Start(ctx context.Context) {
	s.startSessionCleanup(ctx)
	s.startServerMonitor(ctx)
}
