// This file exposes public runtime-state projections consumed by plugin-aware
// frontend shells that need minimal installation and enablement state.

package plugin

import "context"

// RuntimeStateListOutput defines output for public runtime state queries.
type RuntimeStateListOutput struct {
	List []*PluginRuntimeStateItem // List contains public plugin runtime states.
}

// PluginRuntimeStateItem represents public runtime state of one plugin.
type PluginRuntimeStateItem struct {
	Id        string // Id is the stable plugin identifier.
	Installed int    // Installed reports whether the plugin is installed or integrated.
	Enabled   int    // Enabled reports whether the plugin is currently enabled.
	StatusKey string // StatusKey is the host config key used by the public shell.
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
