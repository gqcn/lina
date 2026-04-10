// This file builds review-friendly runtime wasm artifacts from clear-text plugin
// source directories so sample dynamic plugins no longer need to commit binary output.

package plugin

import (
	"encoding/base64"
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

// RuntimeBuildOutput contains the generated runtime artifact bytes and output path.
type RuntimeBuildOutput struct {
	ArtifactPath string
	Content      []byte
	Manifest     *pluginManifest
}

// BuildRuntimeWasmArtifactFromSource builds one runtime wasm artifact from the plugin source tree.
func (s *Service) BuildRuntimeWasmArtifactFromSource(pluginDir string) (*RuntimeBuildOutput, error) {
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{}
	if err := s.loadPluginYAMLFile(manifestPath, manifest); err != nil {
		return nil, gerror.Wrap(err, "failed to load dynamic plugin manifest")
	}
	if err := validateRuntimeBuildManifest(manifest, manifestPath); err != nil {
		return nil, err
	}

	frontendAssets, err := s.collectRuntimeBuilderFrontendAssets(pluginDir)
	if err != nil {
		return nil, err
	}
	hookSpecs, err := s.collectRuntimeBuilderHookSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	resourceSpecs, err := s.collectRuntimeBuilderResourceSpecs(pluginDir, manifest.ID)
	if err != nil {
		return nil, err
	}
	installSQLAssets, err := s.collectRuntimeBuilderSQLAssets(pluginDir, false)
	if err != nil {
		return nil, err
	}
	uninstallSQLAssets, err := s.collectRuntimeBuilderSQLAssets(pluginDir, true)
	if err != nil {
		return nil, err
	}

	content, err := buildRuntimeWasmArtifactContent(
		manifest,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		hookSpecs,
		resourceSpecs,
	)
	if err != nil {
		return nil, err
	}

	return &RuntimeBuildOutput{
		ArtifactPath: filepath.Join(pluginDir, buildPluginDynamicBuildOutputRelativePath(manifest.ID)),
		Content:      content,
		Manifest:     manifest,
	}, nil
}

// WriteRuntimeWasmArtifactFromSource generates and writes one runtime wasm artifact into temp/<plugin-id>.wasm.
func (s *Service) WriteRuntimeWasmArtifactFromSource(pluginDir string) (*RuntimeBuildOutput, error) {
	out, err := s.BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		return nil, err
	}
	if err = gfile.Mkdir(filepath.Dir(out.ArtifactPath)); err != nil {
		return nil, gerror.Wrap(err, "failed to create runtime artifact directory")
	}
	if err = gfile.PutBytes(out.ArtifactPath, out.Content); err != nil {
		return nil, gerror.Wrap(err, "failed to write runtime artifact")
	}
	return out, nil
}

func buildPluginDynamicBuildOutputRelativePath(pluginID string) string {
	return filepath.Join("temp", buildPluginDynamicArtifactFileName(pluginID))
}

func validateRuntimeBuildManifest(manifest *pluginManifest, manifestPath string) error {
	if manifest == nil {
		return gerror.New("dynamic plugin manifest cannot be nil")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return gerror.Newf("dynamic plugin manifest missing id: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return gerror.Newf("dynamic plugin manifest missing name: %s", manifestPath)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return gerror.Newf("dynamic plugin manifest missing version: %s", manifestPath)
	}
	manifest.Type = normalizePluginType(manifest.Type).String()
	if normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return gerror.Newf("dynamic sample manifest type must be dynamic: %s", manifestPath)
	}
	if !pluginManifestIDPattern.MatchString(manifest.ID) {
		return gerror.Newf("dynamic plugin id must use kebab-case: %s", manifest.ID)
	}
	if err := validatePluginManifestSemanticVersion(manifest.Version); err != nil {
		return gerror.Wrapf(err, "dynamic plugin version is invalid: %s", manifestPath)
	}
	return nil
}

func (s *Service) collectRuntimeBuilderFrontendAssets(pluginDir string) ([]*pluginDynamicArtifactFrontendAsset, error) {
	frontendDir := filepath.Join(pluginDir, "frontend", "pages")
	if !gfile.Exists(frontendDir) || !gfile.IsDir(frontendDir) {
		return []*pluginDynamicArtifactFrontendAsset{}, nil
	}

	files, err := gfile.ScanDirFile(frontendDir, "*", true)
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	assets := make([]*pluginDynamicArtifactFrontendAsset, 0, len(files))
	for _, filePath := range files {
		if gfile.IsDir(filePath) {
			continue
		}
		relativePath, relErr := filepath.Rel(frontendDir, filePath)
		if relErr != nil {
			return nil, relErr
		}
		normalizedPath := filepath.ToSlash(relativePath)
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return nil, readErr
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assets = append(assets, &pluginDynamicArtifactFrontendAsset{
			Path:          normalizedPath,
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			ContentType:   contentType,
			Content:       content,
		})
	}
	return assets, nil
}

func (s *Service) collectRuntimeBuilderSQLAssets(
	pluginDir string,
	uninstall bool,
) ([]*pluginDynamicArtifactSQLAsset, error) {
	relativePaths := s.discoverPluginSQLPaths(pluginDir, uninstall)
	assets := make([]*pluginDynamicArtifactSQLAsset, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		sqlPath := filepath.Join(pluginDir, filepath.FromSlash(relativePath))
		content, err := os.ReadFile(sqlPath)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &pluginDynamicArtifactSQLAsset{
			Key:     filepath.Base(sqlPath),
			Content: strings.TrimSpace(string(content)),
		})
	}
	return assets, nil
}

func (s *Service) collectRuntimeBuilderHookSpecs(
	pluginDir string,
	pluginID string,
) ([]*pluginHookSpec, error) {
	hookDir := filepath.Join(pluginDir, "backend", "hooks")
	if !gfile.Exists(hookDir) || !gfile.IsDir(hookDir) {
		return []*pluginHookSpec{}, nil
	}

	files, err := gfile.ScanDirFile(hookDir, "*.yaml", false)
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	items := make([]*pluginHookSpec, 0, len(files))
	for _, filePath := range files {
		spec := &pluginHookSpec{}
		if err = s.loadPluginYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = s.validatePluginHookSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return clonePluginHookSpecs(items), nil
}

func (s *Service) collectRuntimeBuilderResourceSpecs(
	pluginDir string,
	pluginID string,
) ([]*pluginResourceSpec, error) {
	resourceDir := filepath.Join(pluginDir, "backend", "resources")
	if !gfile.Exists(resourceDir) || !gfile.IsDir(resourceDir) {
		return []*pluginResourceSpec{}, nil
	}

	files, err := gfile.ScanDirFile(resourceDir, "*.yaml", false)
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	items := make([]*pluginResourceSpec, 0, len(files))
	for _, filePath := range files {
		spec := &pluginResourceSpec{}
		if err = s.loadPluginYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = s.validatePluginResourceSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, clonePluginResourceSpec(spec))
	}
	return items, nil
}

func buildRuntimeWasmArtifactContent(
	manifest *pluginManifest,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
	hookSpecs []*pluginHookSpec,
	resourceSpecs []*pluginResourceSpec,
) ([]byte, error) {
	embeddedManifest := &pluginDynamicArtifactManifest{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Type:        pluginTypeDynamic.String(),
		Description: manifest.Description,
	}
	runtimeMetadata := &pluginDynamicArtifactMetadata{
		RuntimeKind:        pluginDynamicKindWasm.String(),
		ABIVersion:         pluginDynamicSupportedABIVersion,
		FrontendAssetCount: len(frontendAssets),
		SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
	}

	manifestPayload, err := json.Marshal(embeddedManifest)
	if err != nil {
		return nil, err
	}
	runtimePayload, err := json.Marshal(runtimeMetadata)
	if err != nil {
		return nil, err
	}

	bytes := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionManifest, manifestPayload)
	bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionDynamic, runtimePayload)

	if len(frontendAssets) > 0 {
		payload, marshalErr := json.Marshal(frontendAssets)
		if marshalErr != nil {
			return nil, marshalErr
		}
		bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionFrontend, payload)
	}
	if len(installSQLAssets) > 0 {
		payload, marshalErr := json.Marshal(installSQLAssets)
		if marshalErr != nil {
			return nil, marshalErr
		}
		bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionInstallSQL, payload)
	}
	if len(uninstallSQLAssets) > 0 {
		payload, marshalErr := json.Marshal(uninstallSQLAssets)
		if marshalErr != nil {
			return nil, marshalErr
		}
		bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionUninstallSQL, payload)
	}
	if len(hookSpecs) > 0 {
		payload, marshalErr := json.Marshal(hookSpecs)
		if marshalErr != nil {
			return nil, marshalErr
		}
		bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionBackendHooks, payload)
	}
	if len(resourceSpecs) > 0 {
		payload, marshalErr := json.Marshal(resourceSpecs)
		if marshalErr != nil {
			return nil, marshalErr
		}
		bytes = appendWasmCustomSection(bytes, pluginDynamicWasmSectionBackendRes, payload)
	}
	return bytes, nil
}

func appendWasmCustomSection(content []byte, name string, payload []byte) []byte {
	section := make([]byte, 0, len(name)+len(payload)+8)
	section = appendULEB128(section, uint32(len(name)))
	section = append(section, []byte(name)...)
	section = append(section, payload...)

	content = append(content, 0x00)
	content = appendULEB128(content, uint32(len(section)))
	content = append(content, section...)
	return content
}

func appendULEB128(content []byte, value uint32) []byte {
	current := value
	for {
		part := byte(current & 0x7f)
		current >>= 7
		if current != 0 {
			part |= 0x80
		}
		content = append(content, part)
		if current == 0 {
			return content
		}
	}
}
