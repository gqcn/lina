package config

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
)

func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.svc.Export(ctx, sysconfig.ExportInput{
		Name:      req.Name,
		Key:       req.Key,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=configs.xlsx"))
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
