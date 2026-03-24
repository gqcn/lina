package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
)

// Delete 删除部门
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, c.deptSvc.Delete(ctx, req.Id)
}
