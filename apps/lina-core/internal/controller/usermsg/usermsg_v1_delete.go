package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// Delete 删除用户消息
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.usermsgSvc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
