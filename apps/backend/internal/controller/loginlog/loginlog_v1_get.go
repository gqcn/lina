package loginlog

import (
	"context"

	"backend/api/loginlog/v1"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	record, err := c.loginLogSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{SysLoginLog: record}, nil
}
