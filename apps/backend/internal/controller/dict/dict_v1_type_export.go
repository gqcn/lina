package dict

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "backend/api/dict/v1"
	dictsvc "backend/internal/service/dict"
)

func (c *ControllerV1) TypeExport(ctx context.Context, req *v1.TypeExportReq) (res *v1.TypeExportRes, err error) {
	data, err := c.dictSvc.Export(ctx, dictsvc.ExportInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=dict_types.xlsx"))
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
