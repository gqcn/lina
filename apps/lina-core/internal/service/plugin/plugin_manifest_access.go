// This file separates mutable discovery manifests from the currently active
// manifests so staged dynamic uploads do not immediately replace the release
// that the host is still serving.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// getDesiredPluginManifestByID returns the latest discovered manifest. For
// dynamic plugins this is the mutable staging artifact stored at
// plugin.dynamic.storagePath/<plugin-id>.wasm.
func (s *Service) getDesiredPluginManifestByID(pluginID string) (*pluginManifest, error) {
	if pluginID == "" {
		return nil, gerror.New("插件ID不能为空")
	}
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			return manifest, nil
		}
	}
	return nil, gerror.New("插件不存在")
}

// getActivePluginManifest returns the manifest that is currently effective for
// host serving. Dynamic plugins reload from the archived active release while
// source plugins still come from the compiled-in source workspace.
func (s *Service) getActivePluginManifest(ctx context.Context, pluginID string) (*pluginManifest, error) {
	manifest, err := s.getDesiredPluginManifestByID(pluginID)
	if err != nil {
		return nil, err
	}
	if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return manifest, nil
	}

	registry, err := s.getPluginRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry == nil || registry.Installed != pluginInstalledYes || registry.ReleaseId <= 0 {
		return manifest, nil
	}
	return s.loadActiveDynamicPluginManifest(ctx, registry)
}
