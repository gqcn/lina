package loginlog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/loginlog/v1"
	loginlogsvc "lina-core/internal/service/loginlog"
)

// Export 导出登录日志
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.loginLogSvc.Export(ctx, loginlogsvc.ExportInput{
		UserName:       req.UserName,
		Ip:             req.Ip,
		Status:         req.Status,
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
	r.Response.Header().Set("Content-Disposition", "attachment; filename=login-logs.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
