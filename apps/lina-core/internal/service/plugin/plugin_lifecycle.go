// This file implements runtime plugin install and uninstall flows together with
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
	"lina-core/pkg/pluginhost"
)

// Install executes install lifecycle for a discovered runtime plugin.
func (s *Service) Install(ctx context.Context, pluginID string) error {
	manifest, err := s.getPluginManifestByID(pluginID)
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
		return nil
	}

	// Runtime installation prefers SQL assets embedded in the wasm artifact and
	// falls back to directory-convention SQL only when no embedded bundle exists.
	if err = s.executeManifestSQLFiles(ctx, manifest, pluginMigrationDirectionInstall); err != nil {
		return err
	}
	s.invalidateRuntimeFrontendBundle(ctx, pluginID, "plugin_installed")

	if err = s.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		return err
	}
	registry, err = s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if err = s.syncPluginMetadata(ctx, manifest, registry, "Runtime plugin install lifecycle completed on current node."); err != nil {
		return err
	}
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginInstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
}

// Uninstall executes uninstall lifecycle for an installed runtime plugin.
func (s *Service) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.getPluginManifestByID(pluginID)
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

	if registry.Status == pluginStatusEnabled {
		// Disable before uninstall so routes, menus, and hooks stop participating
		// before any plugin-owned uninstall SQL is executed.
		if err = s.Disable(ctx, pluginID); err != nil {
			return err
		}
	}
	if err = s.executeManifestSQLFiles(ctx, manifest, pluginMigrationDirectionUninstall); err != nil {
		return err
	}
	if err = s.setPluginInstalled(ctx, pluginID, pluginInstalledNo); err != nil {
		return err
	}
	s.invalidateRuntimeFrontendBundle(ctx, pluginID, "plugin_uninstalled")
	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Delete(); err != nil {
		return err
	}
	if err = s.syncPluginNodeState(
		ctx,
		pluginID,
		manifest.Version,
		pluginInstalledNo,
		pluginStatusDisabled,
		"Runtime plugin uninstall lifecycle completed on current node.",
	); err != nil {
		return err
	}
	return s.DispatchHookEvent(
		ctx,
		pluginhost.ExtensionPointPluginUninstalled,
		pluginhost.BuildPluginLifecycleHookPayloadValues(pluginhost.PluginLifecycleHookPayloadInput{
			PluginID: pluginID,
			Name:     manifest.Name,
			Version:  manifest.Version,
		}),
	)
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
	data := do.SysPlugin{
		Installed: installed,
		Status:    pluginStatusDisabled,
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
	return err
}
