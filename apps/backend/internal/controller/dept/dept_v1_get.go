package dept

import (
	"context"

	v1 "backend/api/dept/v1"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	dept, err := c.deptSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{SysDept: dept}, nil
}
