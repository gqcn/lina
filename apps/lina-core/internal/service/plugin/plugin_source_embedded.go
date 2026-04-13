// This file scans source plugins backed by embedded filesystems and resolves
// embedded manifest, SQL, and frontend assets for the plugin service.

package plugin

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginfs"
	"lina-core/pkg/pluginhost"
)

func (s *Service) scanEmbeddedSourcePluginManifests() ([]*pluginManifest, error) {
	sourcePlugins := pluginhost.ListSourcePlugins()
	if len(sourcePlugins) == 0 {
		return []*pluginManifest{}, nil
	}

	sort.Slice(sourcePlugins, func(i int, j int) bool {
		return sourcePlugins[i].ID < sourcePlugins[j].ID
	})

	manifests := make([]*pluginManifest, 0, len(sourcePlugins))
	for _, sourcePlugin := range sourcePlugins {
		if sourcePlugin == nil {
			continue
		}

		embeddedFiles := sourcePlugin.GetEmbeddedFiles()
		if embeddedFiles == nil {
			return nil, gerror.Newf("源码插件缺少内嵌资源声明: %s", sourcePlugin.ID)
		}

		manifestContent, err := fs.ReadFile(embeddedFiles, pluginfs.EmbeddedManifestPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "读取源码插件内嵌清单失败: %s", sourcePlugin.ID)
		}

		manifest := &pluginManifest{
			ManifestPath: pluginfs.BuildEmbeddedManifestPath(sourcePlugin.ID, pluginfs.EmbeddedManifestPath),
			SourcePlugin: sourcePlugin,
		}
		if err = yaml.Unmarshal(manifestContent, manifest); err != nil {
			return nil, gerror.Wrapf(err, "解析源码插件内嵌清单失败: %s", sourcePlugin.ID)
		}
		if err = s.validatePluginManifest(manifest, manifest.ManifestPath); err != nil {
			return nil, err
		}
		if err = s.loadPluginBackendConfig(manifest); err != nil {
			return nil, err
		}

		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

func getSourcePluginEmbeddedFiles(manifest *pluginManifest) fs.FS {
	if manifest == nil || manifest.SourcePlugin == nil {
		return nil
	}
	return manifest.SourcePlugin.GetEmbeddedFiles()
}

func hasSourcePluginEmbeddedFiles(manifest *pluginManifest) bool {
	return getSourcePluginEmbeddedFiles(manifest) != nil
}

func (s *Service) readSourcePluginManifestContent(manifest *pluginManifest) ([]byte, error) {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		content, err := fs.ReadFile(embeddedFiles, pluginfs.EmbeddedManifestPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "读取源码插件内嵌清单失败: %s", manifest.ID)
		}
		return content, nil
	}
	if manifest == nil || strings.TrimSpace(manifest.ManifestPath) == "" {
		return nil, gerror.New("源码插件清单路径不能为空")
	}
	content := gfile.GetBytes(manifest.ManifestPath)
	if len(content) == 0 {
		return nil, gerror.Newf("插件清单为空: %s", manifest.ManifestPath)
	}
	return content, nil
}

func (s *Service) readSourcePluginAssetContent(manifest *pluginManifest, relativePath string) (string, error) {
	normalizedPath, err := pluginfs.NormalizeRelativePath(relativePath)
	if err != nil {
		return "", err
	}

	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		content, err := fs.ReadFile(embeddedFiles, normalizedPath)
		if err != nil {
			return "", gerror.Wrapf(err, "读取源码插件内嵌资源失败: %s", normalizedPath)
		}
		return strings.TrimSpace(string(content)), nil
	}

	sqlPath, err := pluginfs.ResolveResourcePath(manifest.RootDir, normalizedPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(gfile.GetContents(sqlPath)), nil
}

func (s *Service) listPluginInstallSQLPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverSQLPathsFromFS(embeddedFiles, false)
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSQLPaths(manifest.RootDir, false)
}

func (s *Service) listPluginUninstallSQLPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverSQLPathsFromFS(embeddedFiles, true)
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSQLPaths(manifest.RootDir, true)
}

func (s *Service) listPluginFrontendPagePaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverVuePathsFromFS(embeddedFiles, "frontend/pages")
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginPagePaths(manifest.RootDir)
}

func (s *Service) listPluginFrontendSlotPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return pluginfs.DiscoverVuePathsFromFS(embeddedFiles, "frontend/slots")
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSlotPaths(manifest.RootDir)
}
