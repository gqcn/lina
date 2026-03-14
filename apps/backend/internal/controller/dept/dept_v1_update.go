package dept

import (
	"context"

	v1 "backend/api/dept/v1"
	deptsvc "backend/internal/service/dept"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	return nil, c.deptSvc.Update(ctx, deptsvc.UpdateInput{
		Id:       req.Id,
		ParentId: req.ParentId,
		Name:     req.Name,
		OrderNum: req.OrderNum,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   req.Status,
		Remark:   req.Remark,
	})
}
