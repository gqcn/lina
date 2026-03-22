package cron

import (
	"context"

	"lina-core/internal/service/config"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"
)

// Service manages all scheduled/cron tasks.
type Service struct {
	configSvc    *config.Service
	serverMonSvc *servermon.Service
	sessionStore session.Store
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
