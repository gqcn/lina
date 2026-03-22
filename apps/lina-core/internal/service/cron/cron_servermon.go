package cron

import (
	"context"
)

// startServerMonitor starts the server monitor metrics collector.
func (s *Service) startServerMonitor(ctx context.Context) {
	s.serverMonSvc.StartCollector(ctx)
}
