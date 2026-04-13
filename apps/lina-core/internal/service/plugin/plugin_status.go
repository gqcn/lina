// This file synchronizes sys_plugin registry rows and updates install and
// enablement state transitions for discovered plugins.

package plugin

import (
	"context"
	"strings"

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
		stableState := derivePluginHostState(installedState, pluginStatusDisabled)
		data := do.SysPlugin{
			PluginId:     manifest.ID,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Type:         manifest.Type,
			Installed:    installedState,
			Status:       pluginStatusDisabled,
			DesiredState: stableState,
			CurrentState: stableState,
			Generation:   int64(1),
			ManifestPath: manifest.ManifestPath,
			Checksum:     s.buildPluginRegistryChecksum(manifest),
			Remark:       manifest.Description,
		}
		if normalizePluginType(manifest.Type) == pluginTypeSource {
			// Source plugins are compiled into the host, so they appear as already
			// installed and enabled as soon as discovery succeeds.
			data.Status = pluginStatusEnabled
			data.DesiredState = pluginHostStateEnabled.String()
			data.CurrentState = pluginHostStateEnabled.String()
			data.InstalledAt = gtime.Now()
			data.EnabledAt = gtime.Now()
		}

		_, err = dao.SysPlugin.Ctx(ctx).
			Data(data).
			Insert()
		if err != nil {
			return nil, err
		}
		registry, err := s.getPluginRegistry(ctx, manifest.ID)
		if err != nil {
			return nil, err
		}
		if normalizePluginType(manifest.Type) == pluginTypeSource {
			if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
				return nil, err
			}
		}
		if err = s.syncPluginMetadata(ctx, manifest, registry, "Source plugin manifest synchronized into host registry."); err != nil {
			return nil, err
		}
		return s.syncPluginRegistryReleaseReference(ctx, registry, manifest)
	}

	data := do.SysPlugin{
		Name:   manifest.Name,
		Type:   manifest.Type,
		Remark: manifest.Description,
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.buildPluginRegistryChecksum(manifest)
		data.Installed = installedState
		data.DesiredState = derivePluginHostState(installedState, existing.Status)
		data.CurrentState = derivePluginHostState(installedState, existing.Status)
		if existing.Generation <= 0 {
			data.Generation = int64(1)
		}
		if existing.InstalledAt == nil {
			data.InstalledAt = gtime.Now()
		}
		if shouldAutoEnableSourcePlugin(existing) {
			data.Status = pluginStatusEnabled
			data.DesiredState = pluginHostStateEnabled.String()
			data.CurrentState = pluginHostStateEnabled.String()
			data.EnabledAt = gtime.Now()
		}
	} else if !shouldTrackStagedDynamicRelease(existing, manifest) {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.buildPluginRegistryChecksum(manifest)
		if existing.DesiredState == "" {
			data.DesiredState = derivePluginHostState(existing.Installed, existing.Status)
		}
		if existing.CurrentState == "" {
			data.CurrentState = derivePluginHostState(existing.Installed, existing.Status)
		}
		if existing.Generation <= 0 {
			data.Generation = int64(1)
		}
	} else {
		// Keep the active registry version stable while a newer dynamic artifact is
		// merely staged in the mutable storage path. The upgrade switch is driven
		// later by the primary-node reconciler.
		data.ManifestPath = existing.ManifestPath
		data.Checksum = existing.Checksum
	}

	_, err = dao.SysPlugin.Ctx(ctx).
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
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		if err = s.syncPluginMenusAndPermissions(ctx, manifest); err != nil {
			return nil, err
		}
	}
	if err = s.syncPluginMetadata(ctx, manifest, registry, "Source plugin manifest synchronized into host registry."); err != nil {
		return nil, err
	}
	return s.syncPluginRegistryReleaseReference(ctx, registry, manifest)
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

func (s *Service) syncPluginRegistryReleaseReference(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
) (*entity.SysPlugin, error) {
	if registry == nil || manifest == nil {
		return registry, nil
	}
	if strings.TrimSpace(registry.Version) != strings.TrimSpace(manifest.Version) {
		return registry, nil
	}

	release, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil || registry.ReleaseId == release.Id {
		return registry, nil
	}

	_, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{ReleaseId: release.Id}).
		Update()
	if err != nil {
		return nil, err
	}
	return s.getPluginRegistry(ctx, registry.PluginId)
}

// setPluginStatus updates plugin enabled status in sys_plugin.
func (s *Service) setPluginStatus(ctx context.Context, pluginID string, enabled int) error {
	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	installed := pluginInstalledYes
	if registry != nil {
		installed = registry.Installed
	}
	stableState := derivePluginHostState(installed, enabled)
	data := do.SysPlugin{
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if enabled == pluginStatusEnabled {
		data.EnabledAt = gtime.Now()
	} else {
		data.DisabledAt = gtime.Now()
	}

	_, err = dao.SysPlugin.Ctx(ctx).
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

	registry, err = s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil {
		return nil
	}
	if err = s.syncPluginReleaseRuntimeState(ctx, registry); err != nil {
		return err
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
