package post

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// Export 导出岗位
func (c *ControllerV1) Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error) {
	data, err := c.postSvc.Export(ctx, postsvc.ExportInput{
		DeptId: req.DeptId,
		Code:   req.Code,
		Name:   req.Name,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=posts.xlsx"))
	r.Response.WriteOver(data)
	r.ExitAll()
	return nil, nil
}
