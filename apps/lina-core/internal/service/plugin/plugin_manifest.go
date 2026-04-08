package plugin

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"
)

var pluginSQLFileNamePattern = regexp.MustCompile(`^\d{3}-[a-z0-9-]+\.sql$`)

// pluginManifest defines plugin metadata loaded from plugin.yaml.
type pluginManifest struct {
	ID               string                  `yaml:"id"`
	Name             string                  `yaml:"name"`
	Version          string                  `yaml:"version"`
	Type             string                  `yaml:"type"`
	Entry            string                  `yaml:"entry"`
	Description      string                  `yaml:"description"`
	Resources        pluginManifestResources `yaml:"resources"`
	ManifestPath     string
	RootDir          string
	Hooks            []*pluginHookSpec
	BackendResources map[string]*pluginResourceSpec
}

// pluginManifestResources defines manifest-declared resource files.
type pluginManifestResources struct {
	SQL pluginManifestSQL `yaml:"sql"`
}

// pluginManifestSQL defines install/uninstall SQL files in manifest.
type pluginManifestSQL struct {
	Install   []string `yaml:"install"`
	Uninstall []string `yaml:"uninstall"`
}

// scanPluginManifests scans source plugins from apps/lina-plugins and parses plugin.yaml.
func (s *Service) scanPluginManifests() ([]*pluginManifest, error) {
	pluginRootDir, err := s.resolvePluginRootDir()
	if err != nil {
		return []*pluginManifest{}, nil
	}

	manifestFiles, err := gfile.ScanDirFile(pluginRootDir, "plugin.yaml", true)
	if err != nil {
		return nil, err
	}

	manifests := make([]*pluginManifest, 0, len(manifestFiles))
	for _, manifestFile := range manifestFiles {
		content := gfile.GetBytes(manifestFile)
		if len(content) == 0 {
			return nil, gerror.Newf("插件清单为空: %s", manifestFile)
		}

		manifest := &pluginManifest{}
		if err = yaml.Unmarshal(content, manifest); err != nil {
			return nil, gerror.Wrapf(err, "解析插件清单失败: %s", manifestFile)
		}
		if err = s.validatePluginManifest(manifest, manifestFile); err != nil {
			return nil, err
		}
		manifest.ManifestPath = manifestFile
		manifest.RootDir = filepath.Dir(manifestFile)
		if err = s.loadPluginBackendConfig(manifest); err != nil {
			return nil, err
		}

		manifests = append(manifests, manifest)
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})
	return manifests, nil
}

// resolvePluginRootDir resolves plugin root directory from current working directory.
func (s *Service) resolvePluginRootDir() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	candidateDirs := []string{
		filepath.Join(workingDir, "apps", "lina-plugins"),
		filepath.Join(workingDir, "..", "lina-plugins"),
		filepath.Join(workingDir, "..", "..", "lina-plugins"),
	}

	for _, dir := range candidateDirs {
		cleanPath := filepath.Clean(dir)
		if gfile.Exists(cleanPath) && gfile.IsDir(cleanPath) {
			return cleanPath, nil
		}
	}

	return "", gerror.Newf("未找到插件目录，候选路径: %s", strings.Join(candidateDirs, ", "))
}

// validatePluginManifest validates required fields in plugin manifest.
func (s *Service) validatePluginManifest(manifest *pluginManifest, filePath string) error {
	rootDir := filepath.Dir(filePath)

	if manifest.ID == "" {
		return gerror.Newf("插件清单缺少id: %s", filePath)
	}
	if manifest.Name == "" {
		return gerror.Newf("插件清单缺少name: %s", filePath)
	}
	if manifest.Version == "" {
		return gerror.Newf("插件清单缺少version: %s", filePath)
	}
	if manifest.Type == "" {
		manifest.Type = "source"
	}
	if !strings.Contains(manifest.ID, "-") {
		return gerror.Newf("插件ID需使用kebab-case风格: %s", manifest.ID)
	}
	if err := validatePluginSQLPaths(rootDir, manifest.Resources.SQL.Install, false); err != nil {
		return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", filePath)
	}
	if err := validatePluginSQLPaths(rootDir, manifest.Resources.SQL.Uninstall, true); err != nil {
		return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", filePath)
	}
	return nil
}

func validatePluginSQLPaths(rootDir string, relativePaths []string, uninstall bool) error {
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
		if !gfile.Exists(filepath.Join(rootDir, filepath.FromSlash(normalizedPath))) {
			return gerror.Newf("SQL 资源文件不存在: %s", relativePath)
		}
	}

	return nil
}

func validatePluginManifestFilePaths(
	rootDir string,
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
		if !gfile.Exists(filepath.Join(rootDir, filepath.FromSlash(normalizedPath))) {
			return gerror.Newf("插件资源文件不存在: %s", relativePath)
		}
	}

	return nil
}
