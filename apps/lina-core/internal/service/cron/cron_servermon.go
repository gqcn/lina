package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// startServerMonitor starts the server monitor metrics collector.
func (s *Service) startServerMonitor(ctx context.Context) {
	// Collect immediately on startup
	s.serverMonSvc.CollectAndStore(ctx)

	// Then collect periodically via gcron
	cronPattern := fmt.Sprintf("*/%d * * * * *", s.monCfg.IntervalSeconds)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		s.serverMonSvc.CollectAndStore(ctx)
	}, CronServerMonitorCollector)
	if err != nil {
		g.Log().Panicf(ctx, "failed to start server monitor cron: %v", err)
	}
}
