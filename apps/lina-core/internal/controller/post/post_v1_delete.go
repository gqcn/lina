package post

import (
	"context"

	v1 "lina-core/api/post/v1"
)

// Delete 删除岗位
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.postSvc.Delete(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
