package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// startSessionCleanup registers the session cleanup cron job.
// This is a Master-Only job, only executed on the leader node.
func (s *Service) startSessionCleanup(ctx context.Context) {
	cronPattern := fmt.Sprintf("# */%d * * * *", s.sessionCfg.CleanupMinute)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		// Check if current node is the leader before executing
		if !s.IsLeader() {
			g.Log().Debug(ctx, "skipping session cleanup on non-leader node")
			return
		}
		cleaned, cleanErr := s.sessionStore.CleanupInactive(ctx, s.sessionCfg.TimeoutHour)
		if cleanErr != nil {
			g.Log().Warningf(ctx, "session cleanup error: %v", cleanErr)
		} else if cleaned > 0 {
			g.Log().Infof(ctx, "session cleanup: removed %d inactive sessions", cleaned)
		}
	}, CronSessionCleanup)
	if err != nil {
		g.Log().Panicf(ctx, "failed to start session cleanup cron: %v", err)
	}
}
