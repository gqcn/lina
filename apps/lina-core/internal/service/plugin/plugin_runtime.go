package plugin

import "context"

// RuntimeStateListOutput defines output for public runtime state queries.
type RuntimeStateListOutput struct {
	List []*PluginRuntimeStateItem // plugin runtime state list
}

// PluginRuntimeStateItem represents public runtime state of one plugin.
type PluginRuntimeStateItem struct {
	Id        string // plugin id
	Installed int    // installed status: 1=installed, 0=not installed
	Enabled   int    // enabled status: 1=enabled, 0=disabled
	StatusKey string // plugin status config key
}

// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
func (s *Service) ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error) {
	out, err := s.SyncAndList(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*PluginRuntimeStateItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &PluginRuntimeStateItem{
			Id:        item.Id,
			Installed: item.Installed,
			Enabled:   item.Enabled,
			StatusKey: item.StatusKey,
		})
	}
	return &RuntimeStateListOutput{List: items}, nil
}
