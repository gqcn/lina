// This file defines the current runtime wasm artifact contract used by Lina.
// It validates embedded manifest metadata, enforces ABI compatibility, and
// exposes review-friendly summaries for governance persistence.

package plugin

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

const (
	pluginRuntimeWasmSectionManifest     = "lina.plugin.manifest"
	pluginRuntimeWasmSectionRuntime      = "lina.plugin.runtime"
	pluginRuntimeWasmSectionFrontend     = "lina.plugin.frontend.assets"
	pluginRuntimeWasmSectionInstallSQL   = "lina.plugin.install.sql"
	pluginRuntimeWasmSectionUninstallSQL = "lina.plugin.uninstall.sql"
	pluginRuntimeWasmSectionBackendHooks = "lina.plugin.backend.hooks"
	pluginRuntimeWasmSectionBackendRes   = "lina.plugin.backend.resources"
	pluginRuntimeSupportedABIVersion     = "v1"
)

// missingRuntimePluginArtifactError marks the "wasm not generated yet" state so
// discovery can keep runtime plugins visible while lifecycle actions stay strict.
type missingRuntimePluginArtifactError struct {
	rootDir      string
	relativePath string
}

func (e *missingRuntimePluginArtifactError) Error() string {
	return fmt.Sprintf("运行时插件目录缺少 %s: %s", e.relativePath, e.rootDir)
}

func buildPluginRuntimeArtifactFileName(pluginID string) string {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return "plugin.wasm"
	}
	return normalizedID + ".wasm"
}

func buildPluginRuntimeArtifactRelativePath(pluginID string) string {
	return filepath.Join("runtime", buildPluginRuntimeArtifactFileName(pluginID))
}

func resolvePluginRuntimeArtifactPath(rootDir string, pluginID string) (string, error) {
	relativePath := filepath.ToSlash(buildPluginRuntimeArtifactRelativePath(pluginID))
	candidatePath := filepath.Join(rootDir, buildPluginRuntimeArtifactRelativePath(pluginID))
	if gfile.Exists(candidatePath) {
		return candidatePath, nil
	}

	legacyPath := filepath.Join(rootDir, "runtime", "plugin.wasm")
	if gfile.Exists(legacyPath) {
		return legacyPath, nil
	}

	return candidatePath, &missingRuntimePluginArtifactError{
		rootDir:      rootDir,
		relativePath: relativePath,
	}
}

func isMissingRuntimePluginArtifactError(err error) bool {
	var target *missingRuntimePluginArtifactError
	return errors.As(err, &target)
}

// pluginRuntimeArtifact describes one validated runtime wasm artifact.
type pluginRuntimeArtifact struct {
	Path               string
	Checksum           string
	RuntimeKind        string
	ABIVersion         string
	FrontendAssetCount int
	SQLAssetCount      int
	Manifest           *pluginRuntimeArtifactManifest
	FrontendAssets     []*pluginRuntimeArtifactFrontendAsset
	InstallSQLAssets   []*pluginRuntimeArtifactSQLAsset
	UninstallSQLAssets []*pluginRuntimeArtifactSQLAsset
	HookSpecs          []*pluginHookSpec
	ResourceSpecs      []*pluginResourceSpec
}

// pluginRuntimeArtifactManifest stores the plugin identity embedded into wasm.
type pluginRuntimeArtifactManifest struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// pluginRuntimeArtifactMetadata stores the host-owned runtime metadata section.
type pluginRuntimeArtifactMetadata struct {
	RuntimeKind        string `json:"runtimeKind" yaml:"runtimeKind"`
	ABIVersion         string `json:"abiVersion" yaml:"abiVersion"`
	FrontendAssetCount int    `json:"frontendAssetCount,omitempty" yaml:"frontendAssetCount,omitempty"`
	SQLAssetCount      int    `json:"sqlAssetCount,omitempty" yaml:"sqlAssetCount,omitempty"`
}

// pluginRuntimeArtifactFrontendAsset stores one embedded frontend static asset.
type pluginRuntimeArtifactFrontendAsset struct {
	Path          string `json:"path" yaml:"path"`
	ContentBase64 string `json:"contentBase64" yaml:"contentBase64"`
	ContentType   string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Content       []byte `json:"-" yaml:"-"`
}

// pluginRuntimeArtifactSQLAsset stores one embedded SQL migration step.
type pluginRuntimeArtifactSQLAsset struct {
	Key     string `json:"key" yaml:"key"`
	Content string `json:"content" yaml:"content"`
}

// validateRuntimePluginArtifact loads and validates the current runtime wasm artifact.
func (s *Service) validateRuntimePluginArtifact(manifest *pluginManifest, rootDir string) error {
	runtimeArtifactPath, err := resolvePluginRuntimeArtifactPath(rootDir, manifest.ID)
	if err != nil {
		return err
	}

	artifact, err := s.parseRuntimeWasmArtifact(runtimeArtifactPath)
	if err != nil {
		return err
	}
	if artifact.Manifest == nil {
		return gerror.Newf("运行时插件缺少嵌入清单: %s", runtimeArtifactPath)
	}

	artifact.Manifest.Type = normalizePluginType(artifact.Manifest.Type).String()
	if normalizePluginType(artifact.Manifest.Type) != pluginTypeRuntime {
		return gerror.Newf("运行时插件嵌入清单类型必须是 runtime: %s", runtimeArtifactPath)
	}
	if manifest.ID != artifact.Manifest.ID {
		return gerror.Newf("运行时插件嵌入清单 ID 与 plugin.yaml 不一致: %s != %s", artifact.Manifest.ID, manifest.ID)
	}
	if manifest.Name != artifact.Manifest.Name {
		return gerror.Newf("运行时插件嵌入清单名称与 plugin.yaml 不一致: %s != %s", artifact.Manifest.Name, manifest.Name)
	}
	if manifest.Version != artifact.Manifest.Version {
		return gerror.Newf("运行时插件嵌入清单版本与 plugin.yaml 不一致: %s != %s", artifact.Manifest.Version, manifest.Version)
	}

	manifest.RuntimeArtifact = artifact
	return nil
}

// ensureRuntimePluginArtifactAvailable keeps lifecycle actions strict even
// though discovery tolerates a missing local build artifact during list/sync.
func (s *Service) ensureRuntimePluginArtifactAvailable(manifest *pluginManifest, actionLabel string) error {
	if manifest == nil {
		return gerror.New("插件清单不能为空")
	}
	if normalizePluginType(manifest.Type) != pluginTypeRuntime {
		return nil
	}
	if manifest.RuntimeArtifact != nil {
		return nil
	}

	if err := s.validateRuntimePluginArtifact(manifest, manifest.RootDir); err != nil {
		if isMissingRuntimePluginArtifactError(err) {
			return gerror.Newf(
				"运行时插件缺少 %s，无法%s；请先执行 make wasm p=%s 生成产物",
				filepath.ToSlash(buildPluginRuntimeArtifactRelativePath(manifest.ID)),
				actionLabel,
				manifest.ID,
			)
		}
		return gerror.Wrapf(err, "运行时插件产物校验失败，无法%s", actionLabel)
	}
	return nil
}

// parseRuntimeWasmArtifact reads one wasm artifact and extracts Lina custom sections.
func (s *Service) parseRuntimeWasmArtifact(filePath string) (*pluginRuntimeArtifact, error) {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return nil, gerror.Newf("运行时插件产物为空: %s", filePath)
	}
	return s.parseRuntimeWasmArtifactContent(filePath, content)
}

// parseRuntimeWasmArtifactContent parses one wasm artifact directly from memory.
func (s *Service) parseRuntimeWasmArtifactContent(filePath string, content []byte) (*pluginRuntimeArtifact, error) {
	sections, err := parseWasmCustomSections(content)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件产物失败: %s", filePath)
	}

	manifestSection, ok := sections[pluginRuntimeWasmSectionManifest]
	if !ok {
		return nil, gerror.Newf("运行时插件缺少自定义节 %s: %s", pluginRuntimeWasmSectionManifest, filePath)
	}
	runtimeSection, ok := sections[pluginRuntimeWasmSectionRuntime]
	if !ok {
		return nil, gerror.Newf("运行时插件缺少自定义节 %s: %s", pluginRuntimeWasmSectionRuntime, filePath)
	}

	embeddedManifest := &pluginRuntimeArtifactManifest{}
	if err = unmarshalRuntimeArtifactSection(manifestSection, embeddedManifest); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件嵌入清单失败: %s", filePath)
	}
	if strings.TrimSpace(embeddedManifest.ID) == "" ||
		strings.TrimSpace(embeddedManifest.Name) == "" ||
		strings.TrimSpace(embeddedManifest.Version) == "" ||
		strings.TrimSpace(embeddedManifest.Type) == "" {
		return nil, gerror.Newf("运行时插件嵌入清单缺少必填字段: %s", filePath)
	}

	runtimeMetadata := &pluginRuntimeArtifactMetadata{}
	if err = unmarshalRuntimeArtifactSection(runtimeSection, runtimeMetadata); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件运行时元数据失败: %s", filePath)
	}

	frontendAssets, err := parseRuntimeArtifactFrontendAssets(
		filePath,
		sections,
		pluginRuntimeWasmSectionFrontend,
	)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := parseRuntimeArtifactSQLAssets(
		filePath,
		sections,
		pluginRuntimeWasmSectionInstallSQL,
	)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := parseRuntimeArtifactSQLAssets(
		filePath,
		sections,
		pluginRuntimeWasmSectionUninstallSQL,
	)
	if err != nil {
		return nil, err
	}
	hookSpecs, err := s.parseRuntimeArtifactHookSpecs(filePath, embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}
	resourceSpecs, err := s.parseRuntimeArtifactResourceSpecs(filePath, embeddedManifest.ID, sections)
	if err != nil {
		return nil, err
	}

	runtimeKind := strings.TrimSpace(strings.ToLower(runtimeMetadata.RuntimeKind))
	if runtimeKind == "" {
		runtimeKind = pluginRuntimeKindWasm.String()
	}
	if runtimeKind != pluginRuntimeKindWasm.String() {
		return nil, gerror.Newf("运行时插件产物类型仅支持 wasm: %s", runtimeKind)
	}

	abiVersion := strings.TrimSpace(strings.ToLower(runtimeMetadata.ABIVersion))
	if abiVersion == "" {
		return nil, gerror.Newf("运行时插件缺少 ABI 版本: %s", filePath)
	}
	if abiVersion != pluginRuntimeSupportedABIVersion {
		return nil, gerror.Newf("运行时插件 ABI 版本不受支持: %s", runtimeMetadata.ABIVersion)
	}

	totalSQLAssetCount := len(installSQLAssets) + len(uninstallSQLAssets)
	if runtimeMetadata.SQLAssetCount > 0 && runtimeMetadata.SQLAssetCount != totalSQLAssetCount {
		return nil, gerror.Newf(
			"运行时插件 SQL 资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.SQLAssetCount,
			totalSQLAssetCount,
		)
	}
	if runtimeMetadata.SQLAssetCount <= 0 {
		runtimeMetadata.SQLAssetCount = totalSQLAssetCount
	}
	if runtimeMetadata.FrontendAssetCount > 0 && runtimeMetadata.FrontendAssetCount != len(frontendAssets) {
		return nil, gerror.Newf(
			"运行时插件前端资源数量与元数据不一致: metadata=%d actual=%d",
			runtimeMetadata.FrontendAssetCount,
			len(frontendAssets),
		)
	}
	if runtimeMetadata.FrontendAssetCount <= 0 {
		runtimeMetadata.FrontendAssetCount = len(frontendAssets)
	}

	return &pluginRuntimeArtifact{
		Path:               filePath,
		Checksum:           fmt.Sprintf("%x", sha256.Sum256(content)),
		RuntimeKind:        runtimeKind,
		ABIVersion:         abiVersion,
		FrontendAssetCount: maxInt(runtimeMetadata.FrontendAssetCount, 0),
		SQLAssetCount:      maxInt(runtimeMetadata.SQLAssetCount, 0),
		Manifest:           embeddedManifest,
		FrontendAssets:     frontendAssets,
		InstallSQLAssets:   installSQLAssets,
		UninstallSQLAssets: uninstallSQLAssets,
		HookSpecs:          hookSpecs,
		ResourceSpecs:      resourceSpecs,
	}, nil
}

// buildPluginRegistryChecksum returns a review-friendly checksum for current plugin source.
func (s *Service) buildPluginRegistryChecksum(manifest *pluginManifest) string {
	if manifest == nil {
		return ""
	}
	if manifest.RuntimeArtifact != nil {
		return manifest.RuntimeArtifact.Checksum
	}
	if strings.TrimSpace(manifest.ManifestPath) == "" {
		return ""
	}

	content := gfile.GetBytes(manifest.ManifestPath)
	if len(content) == 0 {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// buildRuntimeArtifactRemark summarizes runtime wasm metadata for governance review.
func (s *Service) buildRuntimeArtifactRemark(manifest *pluginManifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}

	return fmt.Sprintf(
		"The host validated one %s runtime artifact using ABI %s with %d embedded frontend assets, %d install SQL assets, and %d uninstall SQL assets declared.",
		manifest.RuntimeArtifact.RuntimeKind,
		manifest.RuntimeArtifact.ABIVersion,
		manifest.RuntimeArtifact.FrontendAssetCount,
		len(manifest.RuntimeArtifact.InstallSQLAssets),
		len(manifest.RuntimeArtifact.UninstallSQLAssets),
	)
}

func unmarshalRuntimeArtifactSection(content []byte, target interface{}) error {
	if err := json.Unmarshal(content, target); err == nil {
		return nil
	}
	return gerror.New("运行时插件自定义节仅支持 JSON 编码")
}

func parseWasmCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < 8 {
		return nil, gerror.New("wasm 文件长度不足")
	}
	if string(content[:4]) != "\x00asm" {
		return nil, gerror.New("wasm 文件头非法")
	}
	if content[4] != 0x01 || content[5] != 0x00 || content[6] != 0x00 || content[7] != 0x00 {
		return nil, gerror.New("wasm 版本非法")
	}

	sections := make(map[string][]byte)
	cursor := 8
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readWasmULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, gerror.New("wasm 节长度越界")
		}

		if sectionID == 0 {
			nameLength, nameCursor, err := readWasmULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, gerror.New("wasm 自定义节名称越界")
			}

			sectionName := string(content[nameCursor:nameEnd])
			sectionPayload := make([]byte, end-nameEnd)
			copy(sectionPayload, content[nameEnd:end])
			sections[sectionName] = sectionPayload
		}

		cursor = end
	}
	return sections, nil
}

func readWasmULEB128(content []byte, start int) (uint32, int, error) {
	var (
		value uint32
		shift uint
	)

	cursor := start
	for {
		if cursor >= len(content) {
			return 0, cursor, gerror.New("wasm ULEB128 数据越界")
		}
		current := content[cursor]
		cursor++

		value |= uint32(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, cursor, nil
		}

		shift += 7
		if shift > 28 {
			return 0, cursor, gerror.New("wasm ULEB128 数值过大")
		}
	}
}

func maxInt(value int, lowerBound int) int {
	if value < lowerBound {
		return lowerBound
	}
	return value
}

func parseRuntimeArtifactSQLAssets(
	filePath string,
	sections map[string][]byte,
	sectionName string,
) ([]*pluginRuntimeArtifactSQLAsset, error) {
	sectionContent, ok := sections[sectionName]
	if !ok {
		return []*pluginRuntimeArtifactSQLAsset{}, nil
	}

	assets := make([]*pluginRuntimeArtifactSQLAsset, 0)
	if err := json.Unmarshal(sectionContent, &assets); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件 SQL 自定义节失败: %s", filePath)
	}
	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("运行时插件 SQL 自定义节存在空项: %s", filePath)
		}
		asset.Key = strings.TrimSpace(asset.Key)
		asset.Content = strings.TrimSpace(asset.Content)
		if asset.Key == "" || asset.Content == "" {
			return nil, gerror.Newf("运行时插件 SQL 自定义节缺少 key 或 content: %s", filePath)
		}
		if strings.Contains(asset.Key, "/") || strings.Contains(asset.Key, "\\") {
			return nil, gerror.Newf("运行时插件 SQL 资源键不能包含路径分隔符: %s", asset.Key)
		}
		if !pluginSQLFileNamePattern.MatchString(asset.Key) {
			return nil, gerror.Newf("运行时插件 SQL 资源键不符合命名规则: %s", asset.Key)
		}
	}
	return assets, nil
}

func (s *Service) parseRuntimeArtifactHookSpecs(
	filePath string,
	pluginID string,
	sections map[string][]byte,
) ([]*pluginHookSpec, error) {
	content, ok := sections[pluginRuntimeWasmSectionBackendHooks]
	if !ok {
		return []*pluginHookSpec{}, nil
	}

	items := make([]*pluginHookSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件后端 Hook 契约失败: %s", filePath)
	}
	for _, item := range items {
		if err := s.validatePluginHookSpec(pluginID, item, filePath); err != nil {
			return nil, err
		}
	}
	return clonePluginHookSpecs(items), nil
}

func (s *Service) parseRuntimeArtifactResourceSpecs(
	filePath string,
	pluginID string,
	sections map[string][]byte,
) ([]*pluginResourceSpec, error) {
	content, ok := sections[pluginRuntimeWasmSectionBackendRes]
	if !ok {
		return []*pluginResourceSpec{}, nil
	}

	items := make([]*pluginResourceSpec, 0)
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件后端资源契约失败: %s", filePath)
	}
	cloned := make([]*pluginResourceSpec, 0, len(items))
	for _, item := range items {
		if err := s.validatePluginResourceSpec(pluginID, item, filePath); err != nil {
			return nil, err
		}
		cloned = append(cloned, clonePluginResourceSpec(item))
	}
	return cloned, nil
}

func parseRuntimeArtifactFrontendAssets(
	filePath string,
	sections map[string][]byte,
	sectionName string,
) ([]*pluginRuntimeArtifactFrontendAsset, error) {
	content, ok := sections[sectionName]
	if !ok {
		return []*pluginRuntimeArtifactFrontendAsset{}, nil
	}

	assets := make([]*pluginRuntimeArtifactFrontendAsset, 0)
	if err := json.Unmarshal(content, &assets); err != nil {
		return nil, gerror.Wrapf(err, "解析运行时插件前端资源失败: %s", filePath)
	}

	for _, asset := range assets {
		if asset == nil {
			return nil, gerror.Newf("运行时插件前端资源不能为空: %s", filePath)
		}
		asset.Path = normalizeRuntimeFrontendAssetPath(asset.Path)
		if asset.Path == "" {
			return nil, gerror.Newf("运行时插件前端资源路径不能为空: %s", filePath)
		}
		if asset.ContentBase64 == "" {
			return nil, gerror.Newf("运行时插件前端资源内容不能为空: %s", asset.Path)
		}

		decoded, err := base64.StdEncoding.DecodeString(asset.ContentBase64)
		if err != nil {
			return nil, gerror.Wrapf(err, "解析运行时插件前端资源内容失败: %s", asset.Path)
		}
		if len(decoded) == 0 {
			return nil, gerror.Newf("运行时插件前端资源内容不能为空: %s", asset.Path)
		}
		asset.Content = decoded
	}
	return assets, nil
}

func normalizeRuntimeFrontendAssetPath(relativePath string) string {
	normalizedPath := strings.TrimSpace(relativePath)
	normalizedPath = strings.ReplaceAll(normalizedPath, "\\", "/")
	normalizedPath = strings.TrimPrefix(normalizedPath, "/")
	normalizedPath = strings.TrimPrefix(normalizedPath, "./")
	normalizedPath = strings.TrimSpace(normalizedPath)
	if normalizedPath == "" {
		return ""
	}
	normalizedPath = filepath.ToSlash(filepath.Clean(normalizedPath))
	if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, "../") {
		return ""
	}
	return normalizedPath
}
