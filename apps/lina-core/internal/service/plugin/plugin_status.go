// This file synchronizes sys_plugin registry rows and updates install and
// enablement state transitions for discovered plugins.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// syncPluginManifest creates or updates a source plugin registry row.
func (s *Service) syncPluginManifest(ctx context.Context, manifest *pluginManifest) (*entity.SysPlugin, error) {
	installedState := pluginInstalledNo
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		installedState = pluginInstalledYes
	}

	existing, err := s.getPluginRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		data := do.SysPlugin{
			PluginId:     manifest.ID,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Type:         manifest.Type,
			Installed:    installedState,
			Status:       pluginStatusDisabled,
			ManifestPath: manifest.ManifestPath,
			Checksum:     s.buildPluginRegistryChecksum(manifest),
			Remark:       manifest.Description,
		}
		if normalizePluginType(manifest.Type) == pluginTypeSource {
			// Source plugins are compiled into the host, so they appear as already
			// installed and enabled as soon as discovery succeeds.
			data.Status = pluginStatusEnabled
			data.InstalledAt = gtime.Now()
			data.EnabledAt = gtime.Now()
		}

		_, err = withPluginRegistryQueryCache(dao.SysPlugin.Ctx(ctx), manifest.ID, -1).
			Data(data).
			Insert()
		if err != nil {
			return nil, err
		}
		registry, err := s.getPluginRegistry(ctx, manifest.ID)
		if err != nil {
			return nil, err
		}
		if err = s.syncPluginMetadata(ctx, manifest, registry, "Source plugin manifest synchronized into host registry."); err != nil {
			return nil, err
		}
		return registry, nil
	}

	data := do.SysPlugin{
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		ManifestPath: manifest.ManifestPath,
		Checksum:     s.buildPluginRegistryChecksum(manifest),
		Remark:       manifest.Description,
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		data.Installed = installedState
		if existing.InstalledAt == nil {
			data.InstalledAt = gtime.Now()
		}
		if shouldAutoEnableSourcePlugin(existing) {
			data.Status = pluginStatusEnabled
			data.EnabledAt = gtime.Now()
		}
	}

	_, err = withPluginRegistryQueryCache(dao.SysPlugin.Ctx(ctx), manifest.ID, -1).
		Where(do.SysPlugin{PluginId: manifest.ID}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}

	registry, err := s.getPluginRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}
	if err = s.syncPluginMetadata(ctx, manifest, registry, "Source plugin manifest synchronized into host registry."); err != nil {
		return nil, err
	}
	return registry, nil
}

func shouldAutoEnableSourcePlugin(plugin *entity.SysPlugin) bool {
	if plugin == nil {
		return false
	}
	if plugin.Status == pluginStatusEnabled {
		return false
	}
	return plugin.EnabledAt == nil && plugin.DisabledAt == nil
}

// setPluginStatus updates plugin enabled status in sys_plugin.
func (s *Service) setPluginStatus(ctx context.Context, pluginID string, enabled int) error {
	data := do.SysPlugin{
		Status: enabled,
	}
	if enabled == pluginStatusEnabled {
		data.EnabledAt = gtime.Now()
	} else {
		data.DisabledAt = gtime.Now()
	}

	_, err := withPluginRegistryQueryCache(dao.SysPlugin.Ctx(ctx), pluginID, -1).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}

	eventName := pluginhost.ExtensionPointPluginDisabled
	if enabled == pluginStatusEnabled {
		eventName = pluginhost.ExtensionPointPluginEnabled
	}
	if err = s.DispatchHookEvent(
		ctx,
		eventName,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Status:   &enabled,
		}),
	); err != nil {
		return err
	}

	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}
	return s.syncPluginNodeState(
		ctx,
		registry.PluginId,
		registry.Version,
		registry.Installed,
		registry.Status,
		"Plugin status updated from management API.",
	)
}

// buildPluginStatusKey builds display key for plugin status.
func (s *Service) buildPluginStatusKey(pluginID string) string {
	return "sys_plugin.status:" + pluginID
}
