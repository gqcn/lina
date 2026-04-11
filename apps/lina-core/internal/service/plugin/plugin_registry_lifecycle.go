// This file centralizes sys_plugin state mutations used by lifecycle actions
// and the background reconciler so generation, release, and stable state fields
// stay consistent across install, enable, disable, upgrade, and rollback flows.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// updatePluginRegistryDesiredState records the management intent that the
// primary reconciler should eventually converge to.
func (s *Service) updatePluginRegistryDesiredState(
	ctx context.Context,
	pluginID string,
	desiredState pluginHostStateValue,
) error {
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{DesiredState: desiredState.String()}).
		Update()
	return err
}

// markPluginRegistryReconciling marks the host row as entering a transient
// reconciliation window while keeping the requested desired state persisted.
func (s *Service) markPluginRegistryReconciling(
	ctx context.Context,
	registry *entity.SysPlugin,
	desiredState pluginHostStateValue,
) error {
	if registry == nil {
		return nil
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			DesiredState: desiredState.String(),
			CurrentState: pluginHostStateReconciling.String(),
		}).
		Update()
	return err
}

// finalizePluginRegistryState applies one stable lifecycle state together with
// the release pointer and next generation number after a successful switch.
func (s *Service) finalizePluginRegistryState(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
	release *entity.SysPluginRelease,
	installed int,
	enabled int,
) (*entity.SysPlugin, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := derivePluginHostState(installed, enabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       enabled,
		DesiredState: stableState,
		CurrentState: stableState,
		Generation:   nextPluginGeneration(registry),
	}
	if manifest != nil {
		data.Version = manifest.Version
		data.ManifestPath = manifest.ManifestPath
		data.Checksum = s.buildPluginRegistryChecksum(manifest)
	}
	if release != nil {
		data.ReleaseId = release.Id
	}
	if installed == pluginInstalledYes {
		if registry.Installed != pluginInstalledYes {
			data.InstalledAt = gtime.Now()
		}
		if enabled == pluginStatusEnabled {
			data.EnabledAt = gtime.Now()
		} else {
			data.DisabledAt = gtime.Now()
		}
	} else {
		data.Status = pluginStatusDisabled
		data.ReleaseId = 0
		data.DisabledAt = gtime.Now()
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	if err != nil {
		return nil, err
	}
	return s.getPluginRegistry(ctx, registry.PluginId)
}

// restorePluginRegistryStableState clears a transient reconcile marker and
// rewrites desired/current state back to the stable registry flags.
func (s *Service) restorePluginRegistryStableState(
	ctx context.Context,
	registry *entity.SysPlugin,
) (*entity.SysPlugin, error) {
	if registry == nil {
		return nil, nil
	}

	stableState := buildStablePluginHostState(registry)
	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(do.SysPlugin{
			DesiredState: stableState,
			CurrentState: stableState,
		}).
		Update()
	if err != nil {
		return nil, err
	}
	return s.getPluginRegistry(ctx, registry.PluginId)
}
