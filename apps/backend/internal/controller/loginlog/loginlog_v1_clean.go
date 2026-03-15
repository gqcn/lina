package loginlog

import (
	"context"

	"backend/api/loginlog/v1"
	loginlogsvc "backend/internal/service/loginlog"
)

func (c *ControllerV1) Clean(ctx context.Context, req *v1.CleanReq) (res *v1.CleanRes, err error) {
	deleted, err := c.loginLogSvc.Clean(ctx, loginlogsvc.CleanInput{
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CleanRes{Deleted: deleted}, nil
}
