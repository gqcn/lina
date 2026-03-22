package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// startSessionCleanup registers the session cleanup cron job.
func (s *Service) startSessionCleanup(ctx context.Context) {
	sessionCfg := s.configSvc.GetSession(ctx)
	cronPattern := fmt.Sprintf("*/%d * * * *", sessionCfg.CleanupMinute)
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		cleaned, cleanErr := s.sessionStore.CleanupInactive(ctx, sessionCfg.TimeoutHour)
		if cleanErr != nil {
			g.Log().Warningf(ctx, "session cleanup error: %v", cleanErr)
		} else if cleaned > 0 {
			g.Log().Infof(ctx, "session cleanup: removed %d inactive sessions", cleaned)
		}
	}, "session-cleanup")
	if err != nil {
		g.Log().Warningf(ctx, "failed to start session cleanup cron: %v", err)
	}
}
