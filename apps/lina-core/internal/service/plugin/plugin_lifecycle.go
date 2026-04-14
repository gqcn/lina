// This file exposes lifecycle and status methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// Install executes the install lifecycle for a discovered dynamic plugin.
func (s *Service) Install(ctx context.Context, pluginID string) error {
	return s.lifecycleSvc.Install(ctx, pluginID)
}

// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
func (s *Service) Uninstall(ctx context.Context, pluginID string) error {
	return s.lifecycleSvc.Uninstall(ctx, pluginID)
}

// UpdateStatus updates plugin status, where status is 1=enabled and 0=disabled.
func (s *Service) UpdateStatus(ctx context.Context, pluginID string, status int) error {
	if status != catalog.StatusDisabled && status != catalog.StatusEnabled {
		return gerror.New("插件状态仅支持0或1")
	}
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if status == catalog.StatusEnabled && catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		if err = s.runtimeSvc.EnsureRuntimeArtifactAvailable(manifest, "启用"); err != nil {
			return err
		}
	}
	if err = s.SyncSourcePlugins(ctx); err != nil {
		return err
	}
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	if err != nil {
		return err
	}
	if !installed {
		return gerror.New("插件未安装")
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeDynamic {
		targetState := catalog.HostStateInstalled.String()
		if status == catalog.StatusEnabled {
			targetState = catalog.HostStateEnabled.String()
		}
		return s.runtimeSvc.ReconcileDynamicPluginRequest(ctx, pluginID, targetState)
	}
	return s.catalogSvc.SetPluginStatus(ctx, pluginID, status)
}

// Enable enables the specified plugin.
func (s *Service) Enable(ctx context.Context, pluginID string) error {
	return s.UpdateStatus(ctx, pluginID, catalog.StatusEnabled)
}

// Disable disables the specified plugin.
func (s *Service) Disable(ctx context.Context, pluginID string) error {
	return s.UpdateStatus(ctx, pluginID, catalog.StatusDisabled)
}

// IsInstalled returns whether a plugin is installed.
func (s *Service) IsInstalled(ctx context.Context, pluginID string) bool {
	installed, err := s.runtimeSvc.CheckIsInstalled(ctx, pluginID)
	return err == nil && installed
}

// IsEnabled returns whether a plugin is enabled.
func (s *Service) IsEnabled(ctx context.Context, pluginID string) bool {
	return s.integrationSvc.IsEnabled(ctx, pluginID)
}
