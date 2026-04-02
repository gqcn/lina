package cron

import (
	"context"

	"lina-core/internal/service/config"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"
)

// Cron job name constants.
const (
	CronSessionCleanup         = "session-cleanup"          // Session cleanup job name
	CronServerMonitorCollector = "server-monitor-collector" // Server monitor collector job name
	CronServerMonitorCleanup   = "server-monitor-cleanup"   // Server monitor cleanup job name
)

// Service manages all scheduled/cron tasks.
type Service struct {
	configSvc    *config.Service    // Configuration service
	serverMonSvc *servermon.Service // Server monitor service
	sessionStore session.Store      // Session store
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
	s.startServerMonitorCleanup(ctx)
}
