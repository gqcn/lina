package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobStatus(ctx context.Context, req *v1.JobStatusReq) (res *v1.JobStatusRes, err error) {
	svc := job.New()
	err = svc.UpdateStatus(ctx, req.Id, req.Status)
	return &v1.JobStatusRes{}, err
}
