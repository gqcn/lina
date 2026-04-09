package plugin

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// UploadRuntimePackage uploads one runtime wasm package into the plugin workspace.
func (c *ControllerV1) UploadRuntimePackage(ctx context.Context, req *v1.UploadRuntimePackageReq) (res *v1.UploadRuntimePackageRes, err error) {
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	out, err := c.pluginSvc.UploadRuntimePackage(ctx, &pluginsvc.RuntimeUploadInput{
		File:             uploadFile,
		OverwriteSupport: req.OverwriteSupport == 1,
	})
	if err != nil {
		return nil, err
	}

	return &v1.UploadRuntimePackageRes{
		Id:          out.Id,
		Name:        out.Name,
		Version:     out.Version,
		Type:        out.Type,
		RuntimeKind: out.RuntimeKind,
		RuntimeAbi:  out.RuntimeABI,
		Installed:   out.Installed,
		Enabled:     out.Enabled,
	}, nil
}
