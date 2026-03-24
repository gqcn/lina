package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
	usermsgsvc "lina-core/internal/service/usermsg"
)

// List 查询用户消息列表
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.usermsgSvc.List(ctx, usermsgsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{List: out.List, Total: out.Total}, nil
}
