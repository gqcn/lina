package loginlog

import (
	"context"

	v1 "lina-core/api/loginlog/v1"
)

// Get 获取登录日志详情
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	record, err := c.loginLogSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{SysLoginLog: record}, nil
}
