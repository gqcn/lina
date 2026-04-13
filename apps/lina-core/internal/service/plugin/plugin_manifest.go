// This file scans plugin directories and validates convention-based manifest,
// SQL, page, and slot resources discovered from the plugin workspace.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginfs"
	"lina-core/pkg/pluginhost"
)

// pluginManifest defines plugin metadata loaded from plugin.yaml.
type pluginManifest struct {
	ID               string            `yaml:"id"`
	Name             string            `yaml:"name"`
	Version          string            `yaml:"version"`
	Type             string            `yaml:"type"`
	Description      string            `yaml:"description"`
	Author           string            `yaml:"author"`
	Homepage         string            `yaml:"homepage"`
	License          string            `yaml:"license"`
	Menus            []*pluginMenuSpec `yaml:"menus"`
	ManifestPath     string
	RootDir          string
	Hooks            []*pluginHookSpec
	BackendResources map[string]*pluginResourceSpec
	Routes           []*pluginbridge.RouteContract
	BridgeSpec       *pluginbridge.BridgeSpec
	HostCapabilities map[string]struct{}
	RuntimeArtifact  *pluginDynamicArtifact
	SourcePlugin     *pluginhost.SourcePlugin
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
		return s.scanEmbeddedSourcePluginManifests()
	}

	manifestFiles, err := gfile.ScanDirFile(pluginRootDir, "plugin.yaml", true)
	if err != nil {
		return nil, err
	}
	sort.Strings(manifestFiles)
	if len(manifestFiles) == 0 {
		return s.scanEmbeddedSourcePluginManifests()
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
		if normalizePluginType(manifest.Type) == pluginTypeDynamic {
			continue
		}
		if sourcePlugin, ok := pluginhost.GetSourcePlugin(strings.TrimSpace(manifest.ID)); ok {
			manifest.SourcePlugin = sourcePlugin
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
	if len(manifests) == 0 {
		return s.scanEmbeddedSourcePluginManifests()
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
			return nil, gerror.Wrapf(loadErr, "解析动态插件产物失败: %s", artifactPath)
		}
		if previousPath, ok := seenIDs[manifest.ID]; ok {
			return nil, gerror.Newf(
				"动态插件ID重复: %s 同时出现在 %s 和 %s",
				manifest.ID,
				previousPath,
				artifactPath,
			)
		}
		seenIDs[manifest.ID] = artifactPath
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
		return nil, gerror.Newf("动态插件缺少嵌入清单: %s", artifactPath)
	}

	manifest := &pluginManifest{
		ID:               strings.TrimSpace(artifact.Manifest.ID),
		Name:             strings.TrimSpace(artifact.Manifest.Name),
		Version:          strings.TrimSpace(artifact.Manifest.Version),
		Type:             normalizePluginType(artifact.Manifest.Type).String(),
		Description:      strings.TrimSpace(artifact.Manifest.Description),
		Menus:            artifact.Manifest.Menus,
		ManifestPath:     "",
		RootDir:          filepath.Dir(artifactPath),
		Routes:           artifact.RouteContracts,
		BridgeSpec:       artifact.BridgeSpec,
		HostCapabilities: pluginbridge.CapabilitySliceToMap(artifact.Capabilities),
		RuntimeArtifact:  artifact,
	}
	if err = s.validateUploadedRuntimeManifest(manifest); err != nil {
		return nil, gerror.Wrapf(err, "动态插件嵌入清单不合法: %s", artifactPath)
	}
	artifact.Manifest.Type = manifest.Type
	// Runtime manifests are reloaded from both the mutable staging artifact and
	// archived active releases. Always hydrate embedded backend contracts here so
	// every caller receives a complete runtime manifest with hook/resource specs.
	if err = s.loadPluginBackendConfig(manifest); err != nil {
		return nil, err
	}
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
	storagePath := s.configSvc.GetPluginDynamicStoragePath(ctx)
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
	fileLabel := strings.TrimSpace(filePath)
	if fileLabel == "" {
		fileLabel = strings.TrimSpace(manifest.ManifestPath)
	}
	if fileLabel == "" {
		fileLabel = manifest.ID
	}

	if manifest.ID == "" {
		return gerror.Newf("插件清单缺少id: %s", fileLabel)
	}
	if manifest.Name == "" {
		return gerror.Newf("插件清单缺少name: %s", fileLabel)
	}
	if manifest.Version == "" {
		return gerror.Newf("插件清单缺少version: %s", fileLabel)
	}
	if manifest.Type == "" {
		manifest.Type = pluginTypeSource.String()
	} else {
		manifest.Type = normalizePluginType(manifest.Type).String()
	}
	if !isSupportedPluginType(manifest.Type) {
		return gerror.Newf("插件类型仅支持 source/dynamic: %s", fileLabel)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return gerror.Newf("插件ID需使用kebab-case风格: %s", manifest.ID)
	}
	if err := validatePluginManifestSemanticVersion(manifest.Version); err != nil {
		return gerror.Wrapf(err, "插件版本不合法: %s", fileLabel)
	}
	if err := s.validatePluginManifestMenus(manifest); err != nil {
		return gerror.Wrapf(err, "插件菜单元数据不合法: %s", fileLabel)
	}
	if normalizePluginType(manifest.Type) == pluginTypeSource {
		if manifest.SourcePlugin != nil && strings.TrimSpace(manifest.SourcePlugin.ID) != "" && manifest.ID != manifest.SourcePlugin.ID {
			return gerror.Newf("源码插件嵌入清单 ID 与注册插件 ID 不一致: %s != %s", manifest.ID, manifest.SourcePlugin.ID)
		}
		goModPath := filepath.Join(rootDir, "go.mod")
		if !hasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(goModPath) {
			return gerror.Newf("源码插件目录缺少 go.mod: %s", rootDir)
		}
		backendEntryPath := filepath.Join(rootDir, "backend", "plugin.go")
		if !hasSourcePluginEmbeddedFiles(manifest) && !gfile.Exists(backendEntryPath) {
			return gerror.Newf("源码插件目录缺少 backend/plugin.go: %s", rootDir)
		}
	} else if err := s.validateRuntimePluginArtifact(manifest, rootDir); err != nil {
		// Runtime plugin discovery no longer depends on source-tree build output.
		// This tolerance now only protects local validation flows that stage a
		// manifest beside a not-yet-generated wasm artifact during tests/review.
		if !isMissingRuntimePluginArtifactError(err) {
			return gerror.Wrapf(err, "动态插件产物校验失败: %s", filePath)
		}
	}
	if embeddedFiles := getSourcePluginEmbeddedFiles(manifest); embeddedFiles != nil {
		if err := pluginfs.ValidateSQLPathsFromFS(embeddedFiles, s.listPluginInstallSQLPaths(manifest), false); err != nil {
			return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateSQLPathsFromFS(embeddedFiles, s.listPluginUninstallSQLPaths(manifest), true); err != nil {
			return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateVuePathsFromFS(
			embeddedFiles,
			s.listPluginFrontendPagePaths(manifest),
			"frontend/pages/",
		); err != nil {
			return gerror.Wrapf(err, "插件清单 frontend page 约束不合法: %s", fileLabel)
		}
		if err := pluginfs.ValidateVuePathsFromFS(
			embeddedFiles,
			s.listPluginFrontendSlotPaths(manifest),
			"frontend/slots/",
		); err != nil {
			return gerror.Wrapf(err, "插件清单 frontend slot 约束不合法: %s", fileLabel)
		}
		return nil
	}
	if err := pluginfs.ValidateSQLPaths(rootDir, s.listPluginInstallSQLPaths(manifest), false); err != nil {
		return gerror.Wrapf(err, "插件清单 install SQL 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateSQLPaths(rootDir, s.listPluginUninstallSQLPaths(manifest), true); err != nil {
		return gerror.Wrapf(err, "插件清单 uninstall SQL 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateVuePaths(
		rootDir,
		s.listPluginFrontendPagePaths(manifest),
		"frontend/pages/",
	); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend page 约束不合法: %s", fileLabel)
	}
	if err := pluginfs.ValidateVuePaths(
		rootDir,
		s.listPluginFrontendSlotPaths(manifest),
		"frontend/slots/",
	); err != nil {
		return gerror.Wrapf(err, "插件清单 frontend slot 约束不合法: %s", fileLabel)
	}
	return nil
}

// discoverPluginSQLPaths discovers plugin SQL files by directory convention.
func (s *Service) discoverPluginSQLPaths(rootDir string, uninstall bool) []string {
	return pluginfs.DiscoverSQLPaths(rootDir, uninstall)
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
	return pluginfs.DiscoverVuePaths(rootDir, relativeDir)
}
