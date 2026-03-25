package dict

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// DataExport exports dictionary data to Excel.
func (c *ControllerV1) DataExport(ctx context.Context, req *v1.DataExportReq) (res *v1.DataExportRes, err error) {
	data, err := c.dictSvc.DataExport(ctx, dictsvc.DataExportInput{
		DictType: req.DictType,
		Label:    req.Label,
		Ids:      req.Ids,
	})
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=dict_data.xlsx"))
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
