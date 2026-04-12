// This file implements dynamic plugin install and uninstall flows together with
// shared helpers that resolve plugin-owned resource paths safely.

package plugin

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

// Install executes install lifecycle for a discovered dynamic plugin.
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
			return nil
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

// resolvePluginResourcePath resolves a plugin relative resource path to an absolute path inside plugin root.
func (s *Service) resolvePluginResourcePath(rootDir string, relativePath string) (string, error) {
	if relativePath == "" {
		return "", gerror.New("插件资源路径不能为空")
	}
	fullPath := filepath.Clean(filepath.Join(rootDir, relativePath))
	rootPath := filepath.Clean(rootDir)
	if fullPath != rootPath && !strings.HasPrefix(fullPath, rootPath+string(filepath.Separator)) {
		return "", gerror.Newf("插件资源路径越界: %s", relativePath)
	}
	if !gfile.Exists(fullPath) {
		return "", gerror.Newf("插件资源文件不存在: %s", fullPath)
	}
	return fullPath, nil
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
