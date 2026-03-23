package cron

import (
	"context"
	"fmt"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/job"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

func init() {
	job.RegisterSystemHandler("session.Cleanup", sessionCleanupHandler)
	job.RegisterSystemHandler("servermon.Collect", servermonCollectHandler)
}

func sessionCleanupHandler(ctx context.Context) error {
	store := session.NewDBStore()
	_, err := store.CleanupInactive(ctx, 24)
	return err
}

func servermonCollectHandler(ctx context.Context) error {
	servermon.New().Collect(ctx)
	return nil
}

func (s *Service) startDynamicJobs(ctx context.Context) {
	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).Where("status", 1).Scan(&jobs)
	if err != nil {
		g.Log().Error(ctx, "load jobs failed:", err)
		return
	}

	for _, j := range jobs {
		if err := s.RegisterJob(ctx, j); err != nil {
			g.Log().Errorf(ctx, "register job[%d] failed: %v", j.Id, err)
		}
	}
}

func (s *Service) RegisterJob(ctx context.Context, j *entity.SysJob) error {
	jobName := getJobName(j.Id)
	_, err := gcron.AddSingleton(ctx, j.CronExpr, func(ctx context.Context) {
		if err := s.jobSvc.Execute(ctx, j); err != nil {
			g.Log().Errorf(ctx, "execute job[%d] failed: %v", j.Id, err)
		}
	}, jobName)
	return err
}

func (s *Service) UnregisterJob(ctx context.Context, jobId uint64) {
	gcron.Remove(getJobName(jobId))
}

func (s *Service) ReloadJobs(ctx context.Context) error {
	gcron.Stop()
	s.startDynamicJobs(ctx)
	return nil
}

func getJobName(jobId uint64) string {
	return fmt.Sprintf("job:%d", jobId)
}
