// This file scans plugin directories and validates convention-based manifest,
// SQL, page, and slot resources discovered from the plugin workspace.

package plugin

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	configsvc "lina-core/internal/service/config"
)

var pluginSQLFileNamePattern = regexp.MustCompile(`^\d{3}-[a-z0-9-]+\.sql$`)
var pluginVueFileExts = map[string]struct{}{
	".vue": {},
}

// pluginManifest defines plugin metadata loaded from plugin.yaml.
type pluginManifest struct {
	ID               string `yaml:"id"`
	Name             string `yaml:"name"`
	Version          string `yaml:"version"`
	Type             string `yaml:"type"`
	Description      string `yaml:"description"`
	Author           string `yaml:"author"`
	Homepage         string `yaml:"homepage"`
	License          string `yaml:"license"`
	ManifestPath     string
	RootDir          string
	Hooks            []*pluginHookSpec
	BackendResources map[string]*pluginResourceSpec
	RuntimeArtifact  *pluginRuntimeArtifact
}

// scanPluginManifests merges source-plugin discovery and runtime-wasm discovery
// into one normalized manifest list used by lifecycle and governance services.
func (s *Service) scanPluginManifests() ([]*pluginManifest, error) {
	sourceManifests, err := s.scanSourcePluginManifests()
	if err != nil {
		return nil, err
	}
	runtimeManifests, err := s.scanRuntimePluginManifests(context.Background())
	if err != nil {
		return nil, err
	}

	manifests := make([]*pluginManifest, 0, len(sourceManifests)+len(runtimeManifests))
	seenIDs := make(map[string]string, len(sourceManifests)+len(runtimeManifests))
	for _, items := range [][]*pluginManifest{sourceManifests, runtimeManifests} {
		for _, manifest := range items {
			if manifest == nil {
				continue
			}
			location := s.buildPluginDiscoveryLocation(manifest)
			if previousFile, ok := seenIDs[manifest.ID]; ok {
				return nil, gerror.Newf(
					"插件ID重复: %s 同时出现在 %s 和 %s",
					manifest.ID,
					previousFile,
					location,
				)
			}
			seenIDs[manifest.ID] = location
			manifests = append(manifests, manifest)
		}
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].ID < manifests[j].ID
	})
	return manifests, nil
}

// scanSourcePluginManifests scans source plugins from apps/lina-plugins. Runtime
// sample directories are skipped here because their clear-text plugin.yaml files
// are only build inputs, not runtime discovery sources.
func (s *Service) scanSourcePluginManifests() ([]*pluginManifest, error) {
	pluginRootDir, err := s.resolvePluginRootDir()
	if err != nil {
		return []*pluginManifest{}, nil
	}

	manifestFiles, err := gfile.ScanDirFile(pluginRootDir, "plugin.yaml", true)
	if err != nil {
		return nil, err
	}
	sort.Strings(manifestFiles)

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
		if normalizePluginType(manifest.Type) == pluginTypeRuntime {
			continue
		}
		if err = s.validatePluginManifest(manifest, manifestFile); err != nil {
			return nil, err
		}
		manifest.ManifestPath = manifestFile
		manifest.RootDir = filepath.Dir(manifestFile)
		// Load backend declarations after the manifest passes structural validation so
		// source-plugin resource scanning always starts from a trusted plugin root.
		if err = s.loadPluginBackendConfig(manifest); err != nil {
			return nil, err
		}

		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

// scanRuntimePluginManifests scans the configured runtime wasm storage directory.
// Discovery is intentionally non-recursive so the host does not impose any extra
// outer directory convention beyond dropping .wasm files into storagePath.
func (s *Service) scanRuntimePluginManifests(ctx context.Context) ([]*pluginManifest, error) {
	storageDir, err := s.resolveRuntimePluginStorageDir(ctx)
	if err != nil {
		return nil, err
	}
	if !gfile.Exists(storageDir) || !gfile.IsDir(storageDir) {
		return []*pluginManifest{}, nil
	}

	artifactFiles, err := gfile.ScanDirFile(storageDir, "*.wasm", false)
	if err != nil {
		return nil, err
	}
	sort.Strings(artifactFiles)

	manifests := make([]*pluginManifest, 0, len(artifactFiles))
	seenIDs := make(map[string]string, len(artifactFiles))
	for _, artifactPath := range artifactFiles {
		manifest, loadErr := s.loadRuntimePluginManifestFromArtifact(artifactPath)
		if loadErr != nil {
			return nil, gerror.Wrapf(loadErr, "解析运行时插件产物失败: %s", artifactPath)
		}
		if previousPath, ok := seenIDs[manifest.ID]; ok {
			return nil, gerror.Newf(
				"运行时插件ID重复: %s 同时出现在 %s 和 %s",
				manifest.ID,
				previousPath,
				artifactPath,
			)
		}
		seenIDs[manifest.ID] = artifactPath
		if err = s.loadPluginBackendConfig(manifest); err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

func (s *Service) buildPluginDiscoveryLocation(manifest *pluginManifest) string {
	if manifest == nil {
		return ""
	}
	if manifest.RuntimeArtifact != nil && strings.TrimSpace(manifest.RuntimeArtifact.Path) != "" {
		return manifest.RuntimeArtifact.Path
	}
	if strings.TrimSpace(manifest.ManifestPath) != "" {
		return manifest.ManifestPath
	}
	return manifest.RootDir
}

func (s *Service) loadRuntimePluginManifestFromArtifact(artifactPath string) (*pluginManifest, error) {
	artifact, err := s.parseRuntimeWasmArtifact(artifactPath)
	if err != nil {
		return nil, err
	}
	if artifact.Manifest == nil {
		return nil, gerror.Newf("运行时插件缺少嵌入清单: %s", artifactPath)
	}

	manifest := &pluginManifest{
		ID:              strings.TrimSpace(artifact.Manifest.ID),
		Name:            strings.TrimSpace(artifact.Manifest.Name),
		Version:         strings.TrimSpace(artifact.Manifest.Version),
		Type:            normalizePluginType(artifact.Manifest.Type).String(),
		Description:     strings.TrimSpace(artifact.Manifest.Description),
		ManifestPath:    "",
		RootDir:         filepath.Dir(artifactPath),
		RuntimeArtifact: artifact,
	}
	if err = s.validateUploadedRuntimeManifest(manifest); err != nil {
		return nil, gerror.Wrapf(err, "运行时插件嵌入清单不合法: %s", artifactPath)
	}
	artifact.Manifest.Type = manifest.Type
	return manifest, nil
}

// resolvePluginRootDir resolves plugin root directory from current working directory.
func (s *Service) resolvePluginRootDir() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	repoRoot, err := findRepoRoot(workingDir)
	if err == nil {
		pluginRootDir := filepath.Join(repoRoot, "apps", "lina-plugins")
		if gfile.Exists(pluginRootDir) && gfile.IsDir(pluginRootDir) {
			return pluginRootDir, nil
		}
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

// resolveRuntimePluginStorageDir resolves the configured runtime wasm storage
// directory. Relative paths are anchored at the repository root when available
// so uploads, manual copies, and automated scans all agree on one shared path.
func (s *Service) resolveRuntimePluginStorageDir(ctx context.Context) (string, error) {
	storagePath := configsvc.New().GetPluginRuntimeStoragePath(ctx)
	if filepath.IsAbs(storagePath) {
		return filepath.Clean(storagePath), nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if repoRoot, repoErr := findRepoRoot(workingDir); repoErr == nil {
		return filepath.Clean(filepath.Join(repoRoot, storagePath)), nil
	}
	return filepath.Clean(filepath.Join(workingDir, storagePath)), nil
}

// validatePluginManifest validates required fields in plugin manifest.
func (s *Service) validatePluginManifest(manifest *pluginManifest, filePath string) error {
	rootDir := filepath.Dir(filePath)
	if strings.TrimSpace(filePath) == "" && strings.TrimSpace(manifest.RootDir) != "" {
		rootDir = manifest.RootDir
	}

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
		manifest.Type = pluginTypeSource.String()
	} else {
		manifest.Type = normalizePluginType(manifest.Type).String()
	}
	if !isSupportedPluginType(manifest.Type) {
		return gerror.Newf("插件类型仅支持 source/runtime: %s", filePath)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return gerror.Newf("插件ID需使用kebab-case风格: %s", manifest.ID)
	}
	if err := validatePluginManifestSemanticVersion(manifest.Version); err != nil {
		return gerror.Wrapf(err, "插件版本不合法: %s", filePath)
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		goModPath := filepath.Join(rootDir, "go.mod")
		if !gfile.Exists(goModPath) {
			return gerror.Newf("源码插件目录缺少 go.mod: %s", rootDir)
		}
		backendEntryPath := filepath.Join(rootDir, "backend", "plugin.go")
		if !gfile.Exists(backendEntryPath) {
			return gerror.Newf("源码插件目录缺少 backend/plugin.go: %s", rootDir)
		}
	} else if err := s.validateRuntimePluginArtifact(manifest, rootDir); err != nil {
		// Runtime plugin discovery no longer depends on source-tree build output.
		// This tolerance now only protects local validation flows that stage a
		// manifest beside a not-yet-generated wasm artifact during tests/review.
		if !isMissingRuntimePluginArtifactError(err) {
			return gerror.Wrapf(err, "运行时插件产物校验失败: %s", filePath)
		}
	}
	if err := validatePluginSQLPaths(rootDir, s.discoverPluginSQLPaths(rootDir, false), false); err != nil {
		return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", filePath)
	}
	if err := validatePluginSQLPaths(rootDir, s.discoverPluginSQLPaths(rootDir, true), true); err != nil {
		return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", filePath)
	}
	if err := validatePluginManifestFilePaths(
		rootDir,
		s.discoverPluginPagePaths(rootDir),
		"frontend/pages/",
		pluginVueFileExts,
	); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend page 约束不合法: %s", filePath)
	}
	if err := validatePluginManifestFilePaths(
		rootDir,
		s.discoverPluginSlotPaths(rootDir),
		"frontend/slots/",
		pluginVueFileExts,
	); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend slot 约束不合法: %s", filePath)
	}
	return nil
}

// discoverPluginSQLPaths discovers plugin SQL files by directory convention.
func (s *Service) discoverPluginSQLPaths(rootDir string, uninstall bool) []string {
	var (
		searchDir = filepath.Join(rootDir, "manifest", "sql")
		relPrefix = "manifest/sql"
	)

	if uninstall {
		searchDir = filepath.Join(rootDir, "manifest", "sql", "uninstall")
		relPrefix = "manifest/sql/uninstall"
	}

	if !gfile.Exists(searchDir) || !gfile.IsDir(searchDir) {
		return []string{}
	}

	sqlFiles, err := gfile.ScanDirFile(searchDir, "*.sql", false)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(sqlFiles))
	for _, sqlFile := range sqlFiles {
		// Return normalized relative paths for validation and execution only. These
		// paths are intentionally not persisted into plugin review tables.
		items = append(items, path.Join(relPrefix, filepath.Base(sqlFile)))
	}
	sort.Strings(items)
	return items
}

// discoverPluginPagePaths discovers plugin page source files by directory convention.
func (s *Service) discoverPluginPagePaths(rootDir string) []string {
	return s.discoverPluginVuePaths(rootDir, filepath.Join("frontend", "pages"))
}

// discoverPluginSlotPaths discovers plugin slot source files by directory convention.
func (s *Service) discoverPluginSlotPaths(rootDir string) []string {
	return s.discoverPluginVuePaths(rootDir, filepath.Join("frontend", "slots"))
}

func (s *Service) discoverPluginVuePaths(rootDir string, relativeDir string) []string {
	searchDir := filepath.Join(rootDir, relativeDir)
	if !gfile.Exists(searchDir) || !gfile.IsDir(searchDir) {
		return []string{}
	}

	resourceFiles, err := gfile.ScanDirFile(searchDir, "*.vue", true)
	if err != nil {
		return []string{}
	}

	items := make([]string, 0, len(resourceFiles))
	for _, resourceFile := range resourceFiles {
		relativePath, relErr := filepath.Rel(rootDir, resourceFile)
		if relErr != nil {
			continue
		}
		items = append(items, path.Clean(strings.ReplaceAll(relativePath, "\\", "/")))
	}
	sort.Strings(items)
	return items
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
