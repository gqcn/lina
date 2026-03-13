package user

import (
	"context"

	"backend/api/user/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, c.userSvc.Delete(ctx, req.Id)
}
