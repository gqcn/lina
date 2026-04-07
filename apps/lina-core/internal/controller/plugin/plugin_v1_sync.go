package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// Sync scans source plugins and synchronizes plugin registry metadata.
func (c *ControllerV1) Sync(ctx context.Context, req *v1.SyncReq) (res *v1.SyncRes, err error) {
	_ = req
	out, err := c.pluginSvc.SyncAndList(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SyncRes{Total: out.Total}, nil
}
