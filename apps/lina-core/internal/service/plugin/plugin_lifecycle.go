package plugin

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
	if manifest.Type == "source" {
		return gerror.New("源码插件随宿主编译集成，不支持安装")
	}

	registry, err := s.syncPluginManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry.Installed == pluginInstalledYes {
		return nil
	}

	if err = s.executeManifestSQLFiles(ctx, manifest.RootDir, s.discoverPluginSQLPaths(manifest.RootDir, false)); err != nil {
		return err
	}

	if err = s.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		return err
	}
	return s.DispatchHookEvent(ctx, pluginhost.ExtensionPointPluginInstalled, map[string]interface{}{
		"pluginId": pluginID,
		"name":     manifest.Name,
		"version":  manifest.Version,
	})
}

// Uninstall executes uninstall lifecycle for an installed runtime plugin.
func (s *Service) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.getPluginManifestByID(pluginID)
	if err != nil {
		return err
	}
	if manifest.Type == "source" {
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
		if err = s.Disable(ctx, pluginID); err != nil {
			return err
		}
	}
	if err = s.executeManifestSQLFiles(ctx, manifest.RootDir, s.discoverPluginSQLPaths(manifest.RootDir, true)); err != nil {
		return err
	}
	if err = s.setPluginInstalled(ctx, pluginID, pluginInstalledNo); err != nil {
		return err
	}
	return s.DispatchHookEvent(ctx, pluginhost.ExtensionPointPluginUninstalled, map[string]interface{}{
		"pluginId": pluginID,
		"name":     manifest.Name,
		"version":  manifest.Version,
	})
}

// executeManifestSQLFiles executes plugin manifest SQL files sequentially.
func (s *Service) executeManifestSQLFiles(ctx context.Context, rootDir string, relativePaths []string) error {
	for _, relativePath := range relativePaths {
		sqlPath, err := s.resolvePluginResourcePath(rootDir, relativePath)
		if err != nil {
			return err
		}
		sqlContent := gfile.GetContents(sqlPath)
		if sqlContent == "" {
			return gerror.Newf("插件SQL文件为空: %s", sqlPath)
		}
		if _, err = g.DB().Exec(ctx, sqlContent); err != nil {
			return gerror.Wrapf(err, "执行插件SQL失败: %s", filepath.Base(sqlPath))
		}
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
