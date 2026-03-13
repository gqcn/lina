package user

import (
	"context"

	"backend/api/user/v1"
	usersvc "backend/internal/service/user"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.userSvc.List(ctx, usersvc.ListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Status:         req.Status,
		Phone:          req.Phone,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{
		List:  out.List,
		Total: out.Total,
	}, nil
}
