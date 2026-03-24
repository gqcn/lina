package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// Exclude 查询排除指定节点的部门列表
func (c *ControllerV1) Exclude(ctx context.Context, req *v1.ExcludeReq) (res *v1.ExcludeRes, err error) {
	list, err := c.deptSvc.Exclude(ctx, deptsvc.ExcludeInput{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ExcludeRes{
		List: list,
	}, nil
}
