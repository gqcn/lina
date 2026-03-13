package user

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"backend/api/user/v1"
	usersvc "backend/internal/service/user"
)

func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.userSvc.Export(ctx, usersvc.ExportInput{
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

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=users.xlsx"))
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
