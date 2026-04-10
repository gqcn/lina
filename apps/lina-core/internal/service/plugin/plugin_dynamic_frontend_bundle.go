// This file keeps runtime frontend assets in memory after they are parsed from
// the runtime wasm artifact stored in plugin.dynamic.storagePath. The wasm
// artifact remains the single source of truth, while the in-memory bundle avoids
// extracting files to a workspace directory before the host can serve them.

package plugin

import (
	"bytes"
	"context"
	"io/fs"
	"lina-core/pkg/logger"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// runtimeFrontendBundle stores one dynamic plugin frontend asset set in memory.
type runtimeFrontendBundle struct {
	PluginID     string
	Version      string
	Checksum     string
	ContentTypes map[string]string
	FileSystem   *runtimeFrontendBundleFS
}

// runtimeFrontendBundleFS exposes runtime frontend assets through the standard
// library fs.ReadFile contract so the host can treat wasm-embedded files like a
// read-only virtual filesystem.
type runtimeFrontendBundleFS struct {
	Files map[string]*runtimeFrontendBundleFile
}

// runtimeFrontendBundleFile stores one immutable runtime asset payload.
type runtimeFrontendBundleFile struct {
	Content []byte
}

type runtimeFrontendBundleOpenFile struct {
	name   string
	reader *bytes.Reader
}

type runtimeFrontendBundleFileInfo struct {
	name string
	size int64
}

var runtimeFrontendBundleCache = struct {
	items map[string]*runtimeFrontendBundle
	mu    sync.RWMutex
}{
	items: map[string]*runtimeFrontendBundle{},
}

func buildRuntimeFrontendBundle(manifest *pluginManifest) (*runtimeFrontendBundle, error) {
	if manifest == nil {
		return nil, gerror.New("插件清单不能为空")
	}
	if manifest.RuntimeArtifact == nil {
		return nil, gerror.New("当前动态插件缺少有效产物")
	}
	if len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}

	var (
		contentTypes = make(map[string]string, len(manifest.RuntimeArtifact.FrontendAssets))
		files        = make(map[string]*runtimeFrontendBundleFile, len(manifest.RuntimeArtifact.FrontendAssets))
	)

	for _, asset := range manifest.RuntimeArtifact.FrontendAssets {
		if asset == nil {
			continue
		}
		if asset.Path == "" {
			return nil, gerror.New("当前动态插件前端资源路径不能为空")
		}
		if len(asset.Content) == 0 {
			return nil, gerror.Newf("当前动态插件前端资源内容为空: %s", asset.Path)
		}

		contentType := strings.TrimSpace(asset.ContentType)
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(asset.Path))
		}
		if contentType == "" {
			contentType = http.DetectContentType(asset.Content)
		}

		contentTypes[asset.Path] = contentType
		files[asset.Path] = &runtimeFrontendBundleFile{
			Content: asset.Content,
		}
	}
	if len(files) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}

	return &runtimeFrontendBundle{
		PluginID:     manifest.ID,
		Version:      manifest.Version,
		Checksum:     manifest.RuntimeArtifact.Checksum,
		ContentTypes: contentTypes,
		FileSystem: &runtimeFrontendBundleFS{
			Files: files,
		},
	}, nil
}

func (b *runtimeFrontendBundle) matchesManifest(manifest *pluginManifest) bool {
	if b == nil || manifest == nil || manifest.RuntimeArtifact == nil {
		return false
	}
	if b.PluginID != manifest.ID || b.Version != manifest.Version {
		return false
	}

	checksum := strings.TrimSpace(manifest.RuntimeArtifact.Checksum)
	if checksum == "" {
		return true
	}
	return b.Checksum == checksum
}

func (b *runtimeFrontendBundle) HasAsset(relativePath string) bool {
	if b == nil || b.FileSystem == nil {
		return false
	}

	normalizedPath, err := normalizeRuntimeFrontendRequestedAssetPath(relativePath)
	if err != nil {
		return false
	}
	_, ok := b.FileSystem.Files[normalizedPath]
	return ok
}

func (b *runtimeFrontendBundle) ReadAsset(relativePath string) ([]byte, string, error) {
	if b == nil || b.FileSystem == nil {
		return nil, "", gerror.New("当前动态插件前端资源不可用")
	}

	normalizedPath, err := normalizeRuntimeFrontendRequestedAssetPath(relativePath)
	if err != nil {
		return nil, "", err
	}

	content, err := fs.ReadFile(b.FileSystem, normalizedPath)
	if err != nil {
		return nil, "", gerror.New("当前动态插件前端资源不存在")
	}
	return content, b.ContentTypes[normalizedPath], nil
}

// ReadFile implements fs.ReadFileFS for the in-memory runtime asset bundle.
func (fsys *runtimeFrontendBundleFS) ReadFile(name string) ([]byte, error) {
	if fsys == nil {
		return nil, fs.ErrNotExist
	}

	normalizedPath := normalizeRuntimeFrontendAssetPath(name)
	if normalizedPath == "" {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist}
	}

	file, ok := fsys.Files[normalizedPath]
	if !ok || file == nil || len(file.Content) == 0 {
		return nil, &fs.PathError{Op: "readfile", Path: normalizedPath, Err: fs.ErrNotExist}
	}
	return file.Content, nil
}

// Open implements fs.FS so standard-library helpers such as fs.ReadFile can
// read runtime assets from the in-memory bundle without a physical directory.
func (fsys *runtimeFrontendBundleFS) Open(name string) (fs.File, error) {
	content, err := fsys.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &runtimeFrontendBundleOpenFile{
		name:   name,
		reader: bytes.NewReader(content),
	}, nil
}

func (f *runtimeFrontendBundleOpenFile) Stat() (fs.FileInfo, error) {
	if f == nil || f.reader == nil {
		return nil, fs.ErrInvalid
	}
	return runtimeFrontendBundleFileInfo{
		name: filepath.Base(f.name),
		size: f.reader.Size(),
	}, nil
}

func (f *runtimeFrontendBundleOpenFile) Read(p []byte) (int, error) {
	if f == nil || f.reader == nil {
		return 0, fs.ErrInvalid
	}
	return f.reader.Read(p)
}

func (f *runtimeFrontendBundleOpenFile) Close() error {
	return nil
}

func (fi runtimeFrontendBundleFileInfo) Name() string {
	return fi.name
}

func (fi runtimeFrontendBundleFileInfo) Size() int64 {
	return fi.size
}

func (fi runtimeFrontendBundleFileInfo) Mode() fs.FileMode {
	return 0o444
}

func (fi runtimeFrontendBundleFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi runtimeFrontendBundleFileInfo) IsDir() bool {
	return false
}

func (fi runtimeFrontendBundleFileInfo) Sys() interface{} {
	return nil
}

func normalizeRuntimeFrontendRequestedAssetPath(relativePath string) (string, error) {
	trimmedPath := strings.TrimSpace(relativePath)
	if trimmedPath == "" || trimmedPath == "/" {
		return "index.html", nil
	}

	normalizedPath := normalizeRuntimeFrontendAssetPath(relativePath)
	if normalizedPath == "" {
		return "", gerror.Newf("运行时前端资源路径越界: %s", relativePath)
	}
	return normalizedPath, nil
}

func (s *Service) ensureRuntimeFrontendBundle(ctx context.Context, manifest *pluginManifest) (*runtimeFrontendBundle, error) {
	if manifest == nil {
		return nil, gerror.New("插件清单不能为空")
	}
	if normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}
	if manifest.RuntimeArtifact == nil {
		if err := s.validateRuntimePluginArtifact(manifest, manifest.RootDir); err != nil {
			return nil, err
		}
	}

	cacheKey := strings.TrimSpace(manifest.ID)
	runtimeFrontendBundleCache.mu.RLock()
	cachedBundle := runtimeFrontendBundleCache.items[cacheKey]
	runtimeFrontendBundleCache.mu.RUnlock()
	if cachedBundle != nil && cachedBundle.matchesManifest(manifest) {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache hit plugin=%s version=%s checksum=%s",
			manifest.ID,
			manifest.Version,
			manifest.RuntimeArtifact.Checksum,
		)
		return cachedBundle, nil
	}
	if cachedBundle != nil {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache stale plugin=%s cachedVersion=%s requestedVersion=%s cachedChecksum=%s requestedChecksum=%s",
			manifest.ID,
			cachedBundle.Version,
			manifest.Version,
			cachedBundle.Checksum,
			manifest.RuntimeArtifact.Checksum,
		)
	} else {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache miss plugin=%s version=%s checksum=%s",
			manifest.ID,
			manifest.Version,
			manifest.RuntimeArtifact.Checksum,
		)
	}

	bundle, err := buildRuntimeFrontendBundle(manifest)
	if err != nil {
		return nil, err
	}

	runtimeFrontendBundleCache.mu.Lock()
	defer runtimeFrontendBundleCache.mu.Unlock()

	currentBundle := runtimeFrontendBundleCache.items[cacheKey]
	if currentBundle != nil && currentBundle.matchesManifest(manifest) {
		logger.Debugf(
			ctx,
			"runtime frontend bundle cache filled concurrently plugin=%s version=%s checksum=%s",
			manifest.ID,
			manifest.Version,
			manifest.RuntimeArtifact.Checksum,
		)
		return currentBundle, nil
	}
	runtimeFrontendBundleCache.items[cacheKey] = bundle
	logger.Debugf(
		ctx,
		"runtime frontend bundle cached plugin=%s version=%s checksum=%s assets=%d",
		manifest.ID,
		manifest.Version,
		manifest.RuntimeArtifact.Checksum,
		len(bundle.FileSystem.Files),
	)
	return bundle, nil
}

func (s *Service) invalidateRuntimeFrontendBundle(ctx context.Context, pluginID string, reason string) {
	cacheKey := strings.TrimSpace(pluginID)
	if cacheKey == "" {
		return
	}

	runtimeFrontendBundleCache.mu.Lock()
	defer runtimeFrontendBundleCache.mu.Unlock()

	if _, ok := runtimeFrontendBundleCache.items[cacheKey]; ok {
		logger.Debugf(ctx, "runtime frontend bundle invalidated plugin=%s reason=%s", cacheKey, strings.TrimSpace(reason))
	} else {
		logger.Debugf(ctx, "runtime frontend bundle invalidate skipped plugin=%s reason=%s cache=empty", cacheKey, strings.TrimSpace(reason))
	}
	delete(runtimeFrontendBundleCache.items, cacheKey)
}

func resetRuntimeFrontendBundleCache() {
	runtimeFrontendBundleCache.mu.Lock()
	defer runtimeFrontendBundleCache.mu.Unlock()

	runtimeFrontendBundleCache.items = map[string]*runtimeFrontendBundle{}
}
