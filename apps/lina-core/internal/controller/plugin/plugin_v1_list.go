package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// List scans plugins and returns current synchronized status list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.pluginSvc.List(ctx, pluginsvc.ListInput{
		ID:        req.Id,
		Name:      req.Name,
		Type:      req.Type,
		Status:    req.Status,
		Installed: req.Installed,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PluginItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.PluginItem{
			Id:             item.Id,
			Name:           item.Name,
			Version:        item.Version,
			Type:           item.Type,
			Description:    item.Description,
			ReleaseVersion: item.ReleaseVersion,
			Installed:      item.Installed,
			InstalledAt:    item.InstalledAt,
			Enabled:        item.Enabled,
			LifecycleState: item.LifecycleState,
			NodeState:      item.NodeState,
			ResourceCount:  item.ResourceCount,
			MigrationState: item.MigrationState,
			StatusKey:      item.StatusKey,
			UpdatedAt:      item.UpdatedAt,
		})
	}

	return &v1.ListRes{List: items, Total: out.Total}, nil
}
