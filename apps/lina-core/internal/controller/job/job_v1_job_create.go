package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobCreate(ctx context.Context, req *v1.JobCreateReq) (res *v1.JobCreateRes, err error) {
	svc := job.New()
	err = svc.Create(ctx, &job.CreateInput{
		Name:        req.Name,
		Group:       req.Group,
		Command:     req.Command,
		CronExpr:    req.CronExpr,
		Description: req.Description,
		Status:      req.Status,
		Singleton:   req.Singleton,
		MaxTimes:    req.MaxTimes,
		Remark:      req.Remark,
	})
	return &v1.JobCreateRes{}, err
}
