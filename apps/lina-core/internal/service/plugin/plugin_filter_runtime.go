package plugin

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

type pluginFilterRuntime struct {
	manifests   []*pluginManifest
	enabledByID map[string]bool
}

func (s *Service) buildFilterRuntime(ctx context.Context) (*pluginFilterRuntime, error) {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return nil, err
	}
	return s.buildFilterRuntimeFromManifests(ctx, manifests)
}

func (s *Service) buildFilterRuntimeFromManifests(
	ctx context.Context,
	manifests []*pluginManifest,
) (*pluginFilterRuntime, error) {
	enabledByID, err := s.buildEnabledPluginMap(ctx, manifests)
	if err != nil {
		return nil, err
	}
	return &pluginFilterRuntime{
		manifests:   manifests,
		enabledByID: enabledByID,
	}, nil
}

func (s *Service) buildEnabledPluginMap(
	ctx context.Context,
	manifests []*pluginManifest,
) (map[string]bool, error) {
	enabledByID := make(map[string]bool, len(manifests))
	pluginIDs := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		if _, ok := enabledByID[pluginID]; ok {
			continue
		}
		enabledByID[pluginID] = false
		pluginIDs = append(pluginIDs, pluginID)
	}
	if len(pluginIDs) == 0 {
		return enabledByID, nil
	}

	var registries []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		WhereIn(dao.SysPlugin.Columns().PluginId, pluginIDs).
		Scan(&registries)
	if err != nil {
		return nil, err
	}

	registryByID := make(map[string]*entity.SysPlugin, len(registries))
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		registryByID[strings.TrimSpace(registry.PluginId)] = registry
	}

	for _, pluginID := range pluginIDs {
		registry := registryByID[pluginID]
		if registry == nil {
			continue
		}
		registry, err = s.reconcileRuntimeRegistryArtifactState(ctx, registry)
		if err != nil {
			return nil, err
		}
		enabledByID[pluginID] = registry != nil &&
			registry.Installed == pluginInstalledYes &&
			registry.Status == pluginStatusEnabled
	}

	return enabledByID, nil
}

func (r *pluginFilterRuntime) isEnabled(pluginID string) bool {
	if r == nil {
		return false
	}
	return r.enabledByID[strings.TrimSpace(pluginID)]
}
