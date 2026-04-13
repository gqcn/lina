// This file implements dynamic plugin install and uninstall flows together with
// shared helpers that resolve plugin-owned resource paths safely.

package plugin

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Install executes install lifecycle for a discovered dynamic plugin. Repeated
// installs are treated as idempotent unless the same version needs a refresh.
func (s *Service) Install(ctx context.Context, pluginID string) error {
	manifest, err := s.getDesiredPluginManifestByID(pluginID)
	if err != nil {
		return err
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		return gerror.New("源码插件随宿主编译集成，不支持安装")
	}
	if err = s.ensureRuntimePluginArtifactAvailable(manifest, "安装"); err != nil {
		return err
	}

	registry, err := s.syncPluginManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry.Installed == pluginInstalledYes {
		compareResult, compareErr := compareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return compareErr
		}
		if compareResult < 0 {
			return gerror.New("不支持回退到更低版本，请使用宿主自动回滚结果或重新上传更高版本")
		}
		if compareResult == 0 {
			// Keeping the same version label does not imply the runtime artifact is
			// unchanged; rebuilt Wasm bytes still need to flow through refresh.
			if !s.shouldRefreshInstalledDynamicRelease(ctx, registry, manifest) {
				return nil
			}
		}
	}

	desiredState := pluginHostStateInstalled
	if registry.Installed == pluginInstalledYes && registry.Status == pluginStatusEnabled {
		desiredState = pluginHostStateEnabled
	}
	if err = s.reconcileDynamicPluginRequest(ctx, pluginID, desiredState); err != nil {
		return err
	}
	if !s.isPrimaryNode() {
		return nil
	}
	return nil
}

// shouldRefreshInstalledDynamicRelease decides whether an already installed
// dynamic release should be re-converged even though the semantic version did
// not change. It compares desired checksum, registry checksum, and archived
// release content to detect rebuilt artifacts and stale archives.
func (s *Service) shouldRefreshInstalledDynamicRelease(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *pluginManifest,
) bool {
	if registry == nil || manifest == nil {
		return false
	}
	if normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return false
	}
	if registry.Installed != pluginInstalledYes {
		return false
	}
	if strings.TrimSpace(registry.Checksum) == "" {
		return true
	}
	desiredChecksum := strings.TrimSpace(s.buildPluginRegistryChecksum(manifest))
	if desiredChecksum == "" {
		return true
	}
	if desiredChecksum != strings.TrimSpace(registry.Checksum) {
		return true
	}

	release, err := s.getPluginRegistryRelease(ctx, registry)
	if err != nil || release == nil {
		return true
	}
	packagePath, err := s.resolvePluginReleasePackagePath(ctx, release)
	if err != nil {
		return true
	}
	archivedManifest, err := s.loadRuntimePluginManifestFromArtifact(packagePath)
	if err != nil || archivedManifest == nil {
		return true
	}
	return strings.TrimSpace(s.buildPluginRegistryChecksum(archivedManifest)) != desiredChecksum
}

// Uninstall executes uninstall lifecycle for an installed dynamic plugin.
func (s *Service) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.getDesiredPluginManifestByID(pluginID)
	if err != nil {
		return err
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		return gerror.New("源码插件随宿主编译集成，不支持卸载")
	}

	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != pluginInstalledYes {
		return nil
	}
	if err = s.reconcileDynamicPluginRequest(ctx, pluginID, pluginHostStateUninstalled); err != nil {
		return err
	}
	return nil
}

// setPluginInstalled updates plugin installation state in sys_plugin.
func (s *Service) setPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	stableState := derivePluginHostState(installed, pluginStatusDisabled)
	data := do.SysPlugin{
		Installed:    installed,
		Status:       pluginStatusDisabled,
		DesiredState: stableState,
		CurrentState: stableState,
	}
	if installed == pluginInstalledYes {
		data.InstalledAt = gtime.Now()
	} else {
		data.DisabledAt = gtime.Now()
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(data).
		Update()
	if err != nil {
		return err
	}
	return nil
}
