// This file keeps versioned dynamic-plugin artifacts in a release archive so
// the host can stage a new upload without losing access to the currently active
// release that is still serving in-flight requests and old plugin pages.

package plugin

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/model/entity"
)

// buildPluginDynamicReleaseArtifactRelativePath returns the versioned archive
// location used by sys_plugin_release.package_path for dynamic-plugin artifacts.
func buildPluginDynamicReleaseArtifactRelativePath(pluginID string, version string) string {
	return filepath.ToSlash(
		filepath.Join(
			"releases",
			strings.TrimSpace(pluginID),
			strings.TrimSpace(version),
			buildPluginDynamicArtifactFileName(pluginID),
		),
	)
}

// archiveRuntimePluginReleaseArtifact copies the currently discovered runtime
// artifact into a versioned archive path and returns that stable relative path.
func (s *Service) archiveRuntimePluginReleaseArtifact(ctx context.Context, manifest *pluginManifest) (string, error) {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return "", gerror.New("动态插件归档要求存在有效产物")
	}

	storageDir, err := s.resolveRuntimePluginStorageDir(ctx)
	if err != nil {
		return "", err
	}

	relativePath := buildPluginDynamicReleaseArtifactRelativePath(manifest.ID, manifest.Version)
	targetPath := filepath.Join(storageDir, filepath.FromSlash(relativePath))
	if gfile.Exists(targetPath) {
		return relativePath, nil
	}

	sourcePath := strings.TrimSpace(manifest.RuntimeArtifact.Path)
	if sourcePath == "" {
		return "", gerror.New("动态插件归档缺少产物路径")
	}

	content := gfile.GetBytes(sourcePath)
	if len(content) == 0 {
		return "", gerror.Newf("动态插件归档读取产物失败: %s", sourcePath)
	}
	if err = gfile.Mkdir(filepath.Dir(targetPath)); err != nil {
		return "", gerror.Wrap(err, "创建动态插件 release 归档目录失败")
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return "", gerror.Wrap(err, "写入动态插件 release 归档文件失败")
	}
	return relativePath, nil
}

// resolvePluginReleasePackagePath resolves one persisted release package path
// into an absolute host path. Relative paths are anchored at the runtime
// storage directory so archived releases and staged artifacts share one root.
func (s *Service) resolvePluginReleasePackagePath(ctx context.Context, release *entity.SysPluginRelease) (string, error) {
	if release == nil {
		return "", gerror.New("插件 release 不能为空")
	}

	packagePath := strings.TrimSpace(release.PackagePath)
	if packagePath == "" {
		return "", gerror.Newf("插件 release 缺少 package_path: %s@%s", release.PluginId, release.ReleaseVersion)
	}
	if filepath.IsAbs(packagePath) {
		return filepath.Clean(packagePath), nil
	}

	storageDir, err := s.resolveRuntimePluginStorageDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(storageDir, filepath.FromSlash(packagePath))), nil
}

// loadRuntimePluginManifestFromRelease reloads one dynamic manifest from its
// persisted release archive instead of the mutable staging file.
func (s *Service) loadRuntimePluginManifestFromRelease(ctx context.Context, release *entity.SysPluginRelease) (*pluginManifest, error) {
	if release == nil {
		return nil, gerror.New("插件 release 不能为空")
	}

	packagePath, err := s.resolvePluginReleasePackagePath(ctx, release)
	if err != nil {
		return nil, err
	}
	return s.loadRuntimePluginManifestFromArtifact(packagePath)
}

// loadActiveDynamicPluginManifest returns the currently active dynamic-plugin
// manifest reloaded from the stable release archive.
func (s *Service) loadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*pluginManifest, error) {
	if registry == nil {
		return nil, gerror.New("插件注册记录不能为空")
	}
	if normalizePluginType(registry.Type) != pluginTypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}

	release, err := s.getPluginRegistryRelease(ctx, registry)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("动态插件缺少当前生效 release: %s", registry.PluginId)
	}
	return s.loadRuntimePluginManifestFromRelease(ctx, release)
}
