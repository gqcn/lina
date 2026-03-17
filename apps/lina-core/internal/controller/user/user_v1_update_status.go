package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error) {
	return nil, c.userSvc.UpdateStatus(ctx, req.Id, req.Status)
}
