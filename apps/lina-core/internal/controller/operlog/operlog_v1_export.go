package operlog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/operlog/v1"
	operlogsvc "lina-core/internal/service/operlog"
)

// Export exports operation logs
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.operLogSvc.Export(ctx, operlogsvc.ExportInput{
		Title:          req.Title,
		OperName:       req.OperName,
		OperType:       req.OperType,
		Status:         req.Status,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
		Ids:            req.Ids,
	})
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=oper-logs.xlsx")
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
