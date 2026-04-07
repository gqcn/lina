package cron

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/config"
	"lina-core/internal/service/election"
	pluginsvc "lina-core/internal/service/plugin"
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
	sessionCfg   *config.SessionConfig // Session configuration
	monCfg       *config.MonitorConfig // Monitor configuration
	serverMonSvc *servermon.Service    // Server monitor service
	sessionStore session.Store         // Session store
	electionSvc  *election.Service     // Leader election service
	pluginSvc    *pluginsvc.Service    // Plugin service
}

// New creates and returns a new Service instance.
func New(
	sessionCfg *config.SessionConfig,
	monCfg *config.MonitorConfig,
	sessionStore session.Store,
	electionSvc *election.Service,
) *Service {
	if electionSvc != nil {
		pluginsvc.SetPrimaryNodeChecker(electionSvc.IsLeader)
	}

	return &Service{
		sessionCfg:   sessionCfg,
		monCfg:       monCfg,
		serverMonSvc: servermon.New(),
		sessionStore: sessionStore,
		electionSvc:  electionSvc,
		pluginSvc:    pluginsvc.New(),
	}
}

// Start registers and starts all cron jobs.
func (s *Service) Start(ctx context.Context) {
	// All-Node Jobs: executed on every node
	s.startServerMonitor(ctx)

	// Master-Only Jobs: only executed on the leader node
	s.startSessionCleanup(ctx)
	s.startServerMonitorCleanup(ctx)
	if err := s.pluginSvc.RegisterCrons(ctx); err != nil {
		g.Log().Warningf(ctx, "register plugin cron jobs failed: %v", err)
	}
}

// IsLeader returns whether the current node is the leader.
func (s *Service) IsLeader() bool {
	return s.electionSvc.IsLeader()
}
