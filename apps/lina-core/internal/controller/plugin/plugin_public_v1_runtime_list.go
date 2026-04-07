package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// RuntimeList returns public plugin runtime states for shell slot rendering.
func (c *PublicControllerV1) RuntimeList(ctx context.Context, req *v1.RuntimeListReq) (res *v1.RuntimeListRes, err error) {
	out, err := c.pluginSvc.ListRuntimeStates(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PluginRuntimeItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.PluginRuntimeItem{
			Id:        item.Id,
			Installed: item.Installed,
			Enabled:   item.Enabled,
			StatusKey: item.StatusKey,
		})
	}
	return &v1.RuntimeListRes{List: items}, nil
}
