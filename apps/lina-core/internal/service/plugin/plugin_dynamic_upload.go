// This file handles runtime wasm package uploads and writes validated runtime
// artifacts into the configured runtime storage directory for later discovery,
// installation, and review.

package plugin

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// DynamicUploadInput defines input for uploading a runtime wasm package.
type DynamicUploadInput struct {
	File             *ghttp.UploadFile // File is the uploaded runtime wasm package.
	OverwriteSupport bool              // OverwriteSupport allows replacing a not-installed runtime package.
}

// DynamicUploadOutput defines output for uploading a runtime wasm package.
type DynamicUploadOutput struct {
	Id          string // Id is the plugin identifier embedded in the runtime package.
	Name        string // Name is the display name embedded in the runtime package.
	Version     string // Version is the plugin version embedded in the runtime package.
	Type        string // Type is the normalized top-level plugin type.
	RuntimeKind string // RuntimeKind is the validated runtime artifact kind.
	RuntimeABI  string // RuntimeABI is the validated runtime ABI version.
	Installed   int    // Installed is the synchronized installation status after upload.
	Enabled     int    // Enabled is the synchronized enablement status after upload.
}

// UploadDynamicPackage validates one runtime wasm package and writes it into plugin.dynamic.storagePath.
func (s *Service) UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error) {
	if in == nil || in.File == nil {
		return nil, gerror.New("请上传动态插件文件")
	}

	source, err := in.File.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "打开动态插件文件失败")
	}
	defer source.Close()

	content, err := io.ReadAll(source)
	if err != nil {
		return nil, gerror.Wrap(err, "读取动态插件文件失败")
	}
	if len(content) == 0 {
		return nil, gerror.New("动态插件文件不能为空")
	}
	return s.storeUploadedRuntimePackage(
		ctx,
		normalizeDynamicUploadFilename(in.File.Filename),
		content,
		in.OverwriteSupport,
	)
}

func normalizeDynamicUploadFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "runtime-plugin.wasm"
	}
	if strings.EqualFold(gfile.ExtName(filename), ".wasm") {
		return filename
	}

	// Some browser upload pipelines downgrade the filename to a generic blob
	// name even when the file content is valid WebAssembly. Runtime validation
	// below still checks the wasm header and embedded Lina metadata, so we only
	// normalize the display name here instead of rejecting the upload early.
	return filepath.Base(filename) + ".wasm"
}

func (s *Service) storeUploadedRuntimePackage(
	ctx context.Context,
	filename string,
	content []byte,
	overwriteSupport bool,
) (*DynamicUploadOutput, error) {
	artifact, err := s.parseRuntimeWasmArtifactContent(filename, content)
	if err != nil {
		return nil, err
	}
	if artifact.Manifest == nil {
		return nil, gerror.New("动态插件嵌入清单不能为空")
	}
	manifest := &pluginManifest{
		ID:              strings.TrimSpace(artifact.Manifest.ID),
		Name:            strings.TrimSpace(artifact.Manifest.Name),
		Version:         strings.TrimSpace(artifact.Manifest.Version),
		Type:            normalizePluginType(artifact.Manifest.Type).String(),
		Description:     strings.TrimSpace(artifact.Manifest.Description),
		RuntimeArtifact: artifact,
	}
	if err = s.validateUploadedRuntimeManifest(manifest); err != nil {
		return nil, err
	}

	storageDir, err := s.resolveRuntimePluginStorageDir(ctx)
	if err != nil {
		return nil, err
	}
	targetPath := filepath.Join(storageDir, buildPluginDynamicArtifactFileName(manifest.ID))

	registry, err := s.getPluginRegistry(ctx, manifest.ID)
	if err != nil {
		return nil, err
	}
	registry, err = s.reconcileRuntimeRegistryArtifactState(ctx, registry)
	if err != nil {
		return nil, err
	}
	if registry != nil && normalizePluginType(registry.Type) != pluginTypeDynamic {
		return nil, gerror.New("已存在同名源码插件，不允许上传动态插件覆盖")
	}
	if registry != nil && registry.Installed == pluginInstalledYes {
		return nil, gerror.New("已安装的动态插件暂不支持通过上传覆盖，请先卸载后再重新上传")
	}
	if conflictPath, conflictErr := s.findDuplicateRuntimeArtifactPath(storageDir, manifest.ID, targetPath); conflictErr != nil {
		return nil, conflictErr
	} else if conflictPath != "" {
		return nil, gerror.Newf("动态插件目录存在重复的插件ID %s，请先移除冲突文件: %s", manifest.ID, conflictPath)
	}
	if gfile.Exists(targetPath) && !overwriteSupport {
		return nil, gerror.New("动态插件文件已存在，请开启覆盖后重试")
	}
	if err = gfile.Mkdir(storageDir); err != nil {
		return nil, gerror.Wrap(err, "创建动态插件存储目录失败")
	}

	backupContent := []byte(nil)
	targetExisted := gfile.Exists(targetPath)
	if targetExisted {
		backupContent = gfile.GetBytes(targetPath)
	}
	if err = gfile.PutBytes(targetPath, content); err != nil {
		return nil, gerror.Wrap(err, "写入动态插件产物失败")
	}
	manifest, err = s.loadRuntimePluginManifestFromArtifact(targetPath)
	if err != nil {
		s.restoreUploadedRuntimeArtifact(targetPath, targetExisted, backupContent)
		return nil, err
	}
	s.invalidateRuntimeFrontendBundle(ctx, manifest.ID, "runtime_package_uploaded")

	registry, err = s.syncPluginManifest(ctx, manifest)
	if err != nil {
		s.restoreUploadedRuntimeArtifact(targetPath, targetExisted, backupContent)
		return nil, err
	}

	return &DynamicUploadOutput{
		Id:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Type:        manifest.Type,
		RuntimeKind: manifest.RuntimeArtifact.RuntimeKind,
		RuntimeABI:  manifest.RuntimeArtifact.ABIVersion,
		Installed:   registry.Installed,
		Enabled:     registry.Status,
	}, nil
}

func (s *Service) findDuplicateRuntimeArtifactPath(storageDir string, pluginID string, targetPath string) (string, error) {
	if !gfile.Exists(storageDir) || !gfile.IsDir(storageDir) {
		return "", nil
	}

	artifactFiles, err := gfile.ScanDirFile(storageDir, "*.wasm", false)
	if err != nil {
		return "", err
	}
	for _, artifactPath := range artifactFiles {
		if filepath.Clean(artifactPath) == filepath.Clean(targetPath) {
			continue
		}
		artifact, parseErr := s.parseRuntimeWasmArtifact(artifactPath)
		if parseErr != nil {
			return "", gerror.Wrapf(parseErr, "解析现有动态插件文件失败: %s", artifactPath)
		}
		if artifact.Manifest != nil && strings.TrimSpace(artifact.Manifest.ID) == pluginID {
			return artifactPath, nil
		}
	}
	return "", nil
}

func (s *Service) restoreUploadedRuntimeArtifact(targetPath string, targetExisted bool, backupContent []byte) {
	if targetExisted {
		_ = gfile.PutBytes(targetPath, backupContent)
		return
	}
	_ = gfile.Remove(targetPath)
}

func (s *Service) validateUploadedRuntimeManifest(manifest *pluginManifest) error {
	if manifest == nil {
		return gerror.New("动态插件清单不能为空")
	}
	manifest.Type = normalizePluginType(manifest.Type).String()
	if manifest.Type != pluginTypeDynamic.String() {
		return gerror.New("动态插件类型必须是 dynamic")
	}
	if manifest.ID == "" || !pluginManifestIDPattern.MatchString(manifest.ID) {
		return gerror.New("动态插件 ID 非法")
	}
	if manifest.Name == "" {
		return gerror.New("动态插件名称不能为空")
	}
	if err := validatePluginManifestSemanticVersion(manifest.Version); err != nil {
		return err
	}
	return nil
}
