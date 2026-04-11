// This file exposes public runtime-state projections consumed by plugin-aware
// frontend shells that need minimal installation and enablement state.

package plugin

import (
	"context"
	"strings"
)

// RuntimeStateListOutput defines output for public runtime state queries.
type RuntimeStateListOutput struct {
	List []*PluginDynamicStateItem // List contains public plugin runtime states.
}

// PluginDynamicStateItem represents public runtime state of one plugin.
type PluginDynamicStateItem struct {
	Id         string // Id is the stable plugin identifier.
	Installed  int    // Installed reports whether the plugin is installed or integrated.
	Enabled    int    // Enabled reports whether the plugin is currently enabled.
	Version    string // Version is the currently active plugin version.
	Generation int64  // Generation is the current active plugin generation on the host.
	StatusKey  string // StatusKey is the host config key used by the public shell.
}

// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
func (s *Service) ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error) {
	registries, err := s.listAllPluginRegistries(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*PluginDynamicStateItem, 0, len(registries))
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		if pluginID == "" {
			continue
		}

		installed := registry.Installed
		enabled := registry.Status
		if normalizePluginType(registry.Type) == pluginTypeDynamic {
			exists, _, err := s.hasRuntimeArtifactStorageFile(ctx, pluginID)
			if err != nil {
				return nil, err
			}
			if !exists {
				installed = pluginInstalledNo
				enabled = pluginStatusDisabled
			}
		}

		generation := registry.Generation
		if generation <= 0 {
			generation = 1
		}

		items = append(items, &PluginDynamicStateItem{
			Id:         pluginID,
			Installed:  installed,
			Enabled:    enabled,
			Version:    registry.Version,
			Generation: generation,
			StatusKey:  s.buildPluginStatusKey(pluginID),
		})
	}
	return &RuntimeStateListOutput{List: items}, nil
}
