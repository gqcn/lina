// This file resolves dynamic plugin frontend assets from in-memory bundles that
// are built from the runtime wasm artifact stored in plugin.dynamic.storagePath.
// The host keeps the wasm artifact as the single source of truth and rebuilds
// the bundle cache on startup or on demand after a server restart.

package plugin

import (
	"context"
	"lina-core/pkg/logger"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
type RuntimeFrontendAssetOutput struct {
	Content     []byte // Content is the raw asset body.
	ContentType string // ContentType is the HTTP content type returned to browsers.
}

// PrewarmRuntimeFrontendBundles rebuilds in-memory runtime frontend bundles for
// enabled dynamic plugins during host startup. Request-time resolution still
// keeps a lazy-loading fallback so one failed preload does not block the host.
func (s *Service) PrewarmRuntimeFrontendBundles(ctx context.Context) error {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return err
	}

	logger.Debugf(ctx, "runtime frontend bundle prewarm started manifests=%d", len(manifests))
	failures := make([]string, 0)
	for _, manifest := range manifests {
		if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeDynamic {
			continue
		}
		if manifest.RuntimeArtifact == nil || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
			s.invalidateRuntimeFrontendBundle(ctx, manifest.ID, "no_embedded_frontend_assets")
			continue
		}

		registry, registryErr := s.getPluginRegistry(ctx, manifest.ID)
		if registryErr != nil {
			failures = append(
				failures,
				gerror.Wrapf(registryErr, "预热动态插件前端资源失败: %s", manifest.ID).Error(),
			)
			continue
		}
		if registry == nil || registry.Installed != pluginInstalledYes || registry.Status != pluginStatusEnabled {
			s.invalidateRuntimeFrontendBundle(ctx, manifest.ID, "plugin_not_enabled_during_prewarm")
			continue
		}

		if _, err = s.ensureRuntimeFrontendBundle(ctx, manifest); err != nil {
			failures = append(
				failures,
				gerror.Wrapf(err, "预热动态插件前端资源失败: %s", manifest.ID).Error(),
			)
			logger.Debugf(ctx, "runtime frontend bundle prewarm failed plugin=%s err=%v", manifest.ID, err)
			continue
		}
		logger.Debugf(ctx, "runtime frontend bundle prewarm succeeded plugin=%s version=%s", manifest.ID, manifest.Version)
	}

	if len(failures) > 0 {
		return gerror.New(strings.Join(failures, "; "))
	}
	logger.Debugf(ctx, "runtime frontend bundle prewarm finished")
	return nil
}

// ResolveRuntimeFrontendAsset resolves one enabled dynamic plugin frontend asset for public serving.
func (s *Service) ResolveRuntimeFrontendAsset(
	ctx context.Context,
	pluginID string,
	version string,
	relativePath string,
) (*RuntimeFrontendAssetOutput, error) {
	manifest, err := s.getPluginManifestByID(pluginID)
	if err != nil {
		return nil, err
	}
	if normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return nil, gerror.New("当前插件不是动态插件")
	}
	if manifest.RuntimeArtifact == nil || len(manifest.RuntimeArtifact.FrontendAssets) == 0 {
		return nil, gerror.New("当前动态插件未声明前端资源")
	}
	if strings.TrimSpace(version) == "" || version != manifest.Version {
		return nil, gerror.New("当前动态插件版本不存在或已切换")
	}

	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil || registry.Installed != pluginInstalledYes || registry.Status != pluginStatusEnabled {
		return nil, gerror.New("当前动态插件未启用")
	}

	bundle, err := s.ensureRuntimeFrontendBundle(ctx, manifest)
	if err != nil {
		return nil, err
	}

	content, contentType, err := bundle.ReadAsset(relativePath)
	if err != nil {
		return nil, err
	}
	logger.Debugf(
		ctx,
		"runtime frontend asset resolved plugin=%s version=%s path=%s contentType=%s",
		pluginID,
		version,
		strings.TrimSpace(relativePath),
		contentType,
	)
	return &RuntimeFrontendAssetOutput{
		Content:     content,
		ContentType: contentType,
	}, nil
}

// BuildRuntimeFrontendPublicBaseURL returns the stable public base URL for runtime assets.
func (s *Service) BuildRuntimeFrontendPublicBaseURL(pluginID string, version string) string {
	return "/plugin-assets/" + strings.TrimSpace(pluginID) + "/" + strings.TrimSpace(version) + "/"
}
