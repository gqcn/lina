package cron

import (
	"context"

	"lina-core/internal/service/config"
	"lina-core/internal/service/election"
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
	electionSvc  *election.Service  // Leader election service
}

// New creates and returns a new Service instance.
func New(sessionStore session.Store, electionSvc *election.Service) *Service {
	return &Service{
		configSvc:    config.New(),
		serverMonSvc: servermon.New(),
		sessionStore: sessionStore,
		electionSvc:  electionSvc,
	}
}

// Start registers and starts all cron jobs.
func (s *Service) Start(ctx context.Context) {
	// All-Node Jobs: executed on every node
	s.startServerMonitor(ctx)

	// Master-Only Jobs: only executed on the leader node
	s.startSessionCleanup(ctx)
	s.startServerMonitorCleanup(ctx)
}

// IsLeader returns whether the current node is the leader.
func (s *Service) IsLeader() bool {
	return s.electionSvc.IsLeader()
}

