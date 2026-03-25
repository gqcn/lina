package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// startServerMonitor starts the server monitor metrics collector.
func (s *Service) startServerMonitor(ctx context.Context) {
	monCfg := s.configSvc.GetMonitor(ctx)

	// Collect immediately on startup
	s.serverMonSvc.CollectAndStore(ctx)

	// Then collect periodically via gcron
	cronPattern := fmt.Sprintf("*/%d * * * * *", monCfg.IntervalSeconds)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		s.serverMonSvc.CollectAndStore(ctx)
	}, "server-monitor-collector")
	if err != nil {
		g.Log().Warningf(ctx, "failed to start server monitor cron: %v", err)
	}
}
