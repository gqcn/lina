package plugin

import (
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginhost"
)

const sourcePluginEmbeddedManifestPath = "plugin.yaml"

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

		manifestContent, err := fs.ReadFile(embeddedFiles, sourcePluginEmbeddedManifestPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "读取源码插件内嵌清单失败: %s", sourcePlugin.ID)
		}

		manifest := &pluginManifest{
			ManifestPath: buildEmbeddedSourcePluginManifestPath(sourcePlugin.ID, sourcePluginEmbeddedManifestPath),
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

func buildEmbeddedSourcePluginManifestPath(pluginID string, relativePath string) string {
	normalizedPluginID := strings.TrimSpace(pluginID)
	normalizedPath := path.Clean(strings.TrimSpace(relativePath))
	if normalizedPath == "" || normalizedPath == "." {
		normalizedPath = sourcePluginEmbeddedManifestPath
	}
	if normalizedPluginID == "" {
		return path.Join("embedded", "source-plugins", normalizedPath)
	}
	return path.Join("embedded", "source-plugins", normalizedPluginID, normalizedPath)
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
		content, err := fs.ReadFile(embeddedFiles, sourcePluginEmbeddedManifestPath)
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
	normalizedPath := path.Clean(strings.ReplaceAll(strings.TrimSpace(relativePath), "\\", "/"))
	if normalizedPath == "" || normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
		return "", gerror.Newf("插件资源路径非法: %s", relativePath)
	}

	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		content, err := fs.ReadFile(embeddedFiles, normalizedPath)
		if err != nil {
			return "", gerror.Wrapf(err, "读取源码插件内嵌资源失败: %s", normalizedPath)
		}
		return strings.TrimSpace(string(content)), nil
	}

	sqlPath, err := s.resolvePluginResourcePath(manifest.RootDir, normalizedPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(gfile.GetContents(sqlPath)), nil
}

func (s *Service) listPluginInstallSQLPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return discoverPluginSQLPathsFromFS(embeddedFiles, false)
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSQLPaths(manifest.RootDir, false)
}

func (s *Service) listPluginUninstallSQLPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return discoverPluginSQLPathsFromFS(embeddedFiles, true)
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSQLPaths(manifest.RootDir, true)
}

func (s *Service) listPluginFrontendPagePaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return discoverPluginVuePathsFromFS(embeddedFiles, "frontend/pages")
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginPagePaths(manifest.RootDir)
}

func (s *Service) listPluginFrontendSlotPaths(manifest *pluginManifest) []string {
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		return discoverPluginVuePathsFromFS(embeddedFiles, "frontend/slots")
	}
	if manifest == nil {
		return []string{}
	}
	return s.discoverPluginSlotPaths(manifest.RootDir)
}

func discoverPluginSQLPathsFromFS(fileSystem fs.FS, uninstall bool) []string {
	searchDir := "manifest/sql"
	if uninstall {
		searchDir = "manifest/sql/uninstall"
	}

	entries, err := fs.ReadDir(fileSystem, searchDir)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
			continue
		}
		items = append(items, path.Join(searchDir, entry.Name()))
	}
	sort.Strings(items)
	return items
}

func discoverPluginVuePathsFromFS(fileSystem fs.FS, searchDir string) []string {
	items := make([]string, 0)
	if err := fs.WalkDir(fileSystem, searchDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d == nil || d.IsDir() {
			return walkErr
		}
		if path.Ext(currentPath) != ".vue" {
			return nil
		}
		items = append(items, path.Clean(currentPath))
		return nil
	}); err != nil {
		return []string{}
	}
	sort.Strings(items)
	return items
}

func validatePluginSQLPathsInFS(fileSystem fs.FS, relativePaths []string, uninstall bool) error {
	var (
		expectedDir    = "manifest/sql"
		expectedPrefix = "manifest/sql/"
	)

	if uninstall {
		expectedDir = "manifest/sql/uninstall"
		expectedPrefix = "manifest/sql/uninstall/"
	}

	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("SQL 资源路径不能为空")
		}

		normalizedPath := path.Clean(strings.ReplaceAll(relativePath, "\\", "/"))
		if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
			return gerror.Newf("SQL 资源路径非法: %s", relativePath)
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("SQL 资源路径必须放在 %s: %s", expectedPrefix, relativePath)
		}
		if !uninstall && strings.HasPrefix(normalizedPath, "manifest/sql/uninstall/") {
			return gerror.Newf("安装 SQL 不允许放在 manifest/sql/uninstall/: %s", relativePath)
		}
		if path.Dir(normalizedPath) != expectedDir {
			return gerror.Newf("SQL 资源必须放在 %s 根目录: %s", expectedDir, relativePath)
		}
		if !pluginSQLFileNamePattern.MatchString(path.Base(normalizedPath)) {
			return gerror.Newf("SQL 文件名必须使用 {序号}-{当前迭代名称}.sql: %s", relativePath)
		}
		if _, err := fs.Stat(fileSystem, normalizedPath); err != nil {
			return gerror.Newf("SQL 资源文件不存在: %s", relativePath)
		}
	}

	return nil
}

func validatePluginManifestFilePathsInFS(
	fileSystem fs.FS,
	relativePaths []string,
	expectedPrefix string,
	allowedExt map[string]struct{},
) error {
	for _, relativePath := range relativePaths {
		if relativePath == "" {
			return gerror.New("插件资源路径不能为空")
		}

		normalizedPath := path.Clean(strings.ReplaceAll(relativePath, "\\", "/"))
		if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
			return gerror.Newf("插件资源路径非法: %s", relativePath)
		}
		if !strings.HasPrefix(normalizedPath, expectedPrefix) {
			return gerror.Newf("插件资源路径必须放在 %s 下: %s", expectedPrefix, relativePath)
		}
		if len(allowedExt) > 0 {
			if _, ok := allowedExt[strings.ToLower(path.Ext(normalizedPath))]; !ok {
				return gerror.Newf("插件资源文件类型不支持: %s", relativePath)
			}
		}
		if _, err := fs.Stat(fileSystem, normalizedPath); err != nil {
			return gerror.Newf("插件资源文件不存在: %s", relativePath)
		}
	}

	return nil
}
