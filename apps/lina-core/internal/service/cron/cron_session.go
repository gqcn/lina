package cron

import (
	"context"
	"fmt"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcron"
)

// startSessionCleanup registers the session cleanup cron job.
// This is a Master-Only job, only executed on the leader node.
func (s *Service) startSessionCleanup(ctx context.Context) {
	cronPattern := fmt.Sprintf("# */%d * * * *", s.sessionCfg.CleanupMinute)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		// Check if current node is the leader before executing
		if !s.IsLeader() {
			logger.Debug(ctx, "skipping session cleanup on non-leader node")
			return
		}
		cleaned, cleanErr := s.sessionStore.CleanupInactive(ctx, s.sessionCfg.TimeoutHour)
		if cleanErr != nil {
			logger.Warningf(ctx, "session cleanup error: %v", cleanErr)
		} else if cleaned > 0 {
			logger.Infof(ctx, "session cleanup: removed %d inactive sessions", cleaned)
		}
	}, CronSessionCleanup)
	if err != nil {
		logger.Panicf(ctx, "failed to start session cleanup cron: %v", err)
	}
}
