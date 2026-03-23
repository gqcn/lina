package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobRun(ctx context.Context, req *v1.JobRunReq) (res *v1.JobRunRes, err error) {
	svc := job.New()
	err = svc.Run(ctx, req.Id)
	return &v1.JobRunRes{}, err
}
