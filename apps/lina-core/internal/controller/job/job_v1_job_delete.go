package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobDelete(ctx context.Context, req *v1.JobDeleteReq) (res *v1.JobDeleteRes, err error) {
	svc := job.New()
	err = svc.Delete(ctx, req.Ids)
	return &v1.JobDeleteRes{}, err
}
