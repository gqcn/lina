package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// Install executes plugin install lifecycle.
func (c *ControllerV1) Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error) {
	if err = c.pluginSvc.Install(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.InstallRes{Id: req.Id, Installed: 1, Enabled: 0}, nil
}
