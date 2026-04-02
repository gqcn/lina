package cron

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// startServerMonitorCleanup starts the stale monitor records cleanup job.
// It runs every hour to delete records where updated_at is older than
// (collection_interval * retention_multiplier) seconds.
func (s *Service) startServerMonitorCleanup(ctx context.Context) {
	monCfg := s.configSvc.GetMonitor(ctx)

	// Calculate stale threshold: interval * multiplier
	staleThreshold := time.Duration(monCfg.IntervalSeconds*monCfg.RetentionMultiplier) * time.Second

	_, err := gcron.Add(ctx, "# * * * * *", func(ctx context.Context) {
		cleaned, cleanErr := s.serverMonSvc.CleanupStale(ctx, staleThreshold)
		if cleanErr != nil {
			g.Log().Errorf(ctx, "failed to cleanup stale monitor records: %v", cleanErr)
			return
		}
		if cleaned > 0 {
			g.Log().Infof(
				ctx,
				"cleaned up %d stale monitor records (older than %v)",
				cleaned, time.Now().Add(-staleThreshold).Format("2006-01-02 15:04:05"),
			)
		}
	}, CronServerMonitorCleanup)
	if err != nil {
		g.Log().Warningf(ctx, "failed to start server monitor cleanup cron: %v", err)
	}
}
