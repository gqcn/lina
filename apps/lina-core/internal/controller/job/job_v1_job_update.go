package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobUpdate(ctx context.Context, req *v1.JobUpdateReq) (res *v1.JobUpdateRes, err error) {
	svc := job.New()
	err = svc.Update(ctx, &job.UpdateInput{
		Id:          req.Id,
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
	return &v1.JobUpdateRes{}, err
}
