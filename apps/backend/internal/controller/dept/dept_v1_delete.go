package dept

import (
	"context"

	v1 "backend/api/dept/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, c.deptSvc.Delete(ctx, req.Id)
}
