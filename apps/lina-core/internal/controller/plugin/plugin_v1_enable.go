package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// Enable updates plugin status to enabled.
func (c *ControllerV1) Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error) {
	if err = c.pluginSvc.UpdateStatusWithAuthorization(ctx, req.Id, 1, buildAuthorizationInput(req.Authorization)); err != nil {
		return nil, err
	}

	return &v1.EnableRes{Id: req.Id, Enabled: 1}, nil
}
