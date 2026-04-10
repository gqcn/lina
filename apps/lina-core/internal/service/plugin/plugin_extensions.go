// This file bridges pluginhost callback registrations into host route, cron,
// after-auth, menu-filter, and permission-filter integration flows.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
func (s *Service) RegisterHTTPRoutes(
	ctx context.Context,
	pluginGroup *ghttp.RouterGroup,
	middlewares pluginhost.RouteMiddlewares,
) error {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return err
	}

	checker := s.buildBackgroundEnabledChecker()
	for _, manifest := range manifests {
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		registrar := pluginhost.NewRouteRegistrar(
			pluginGroup,
			manifest.ID,
			checker,
			middlewares,
		)
		for _, handler := range sourcePlugin.GetRouteRegistrars() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			// Route registration happens at host startup, so the host executes every
			// published registrar once while still guarding runtime access by plugin state.
			if err = handler.Handler(ctx, registrar); err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterCrons registers callback-contributed cron jobs for source plugins.
func (s *Service) RegisterCrons(ctx context.Context) error {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return err
	}

	checker := s.buildBackgroundEnabledChecker()
	for _, manifest := range manifests {
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		registrar := pluginhost.NewCronRegistrar(
			manifest.ID,
			checker,
			s.buildPrimaryNodeChecker(),
		)
		for _, handler := range sourcePlugin.GetCronRegistrars() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			if err = handler.Handler(ctx, registrar); err != nil {
				return err
			}
		}
	}
	return nil
}

// DispatchAfterAuthRequest dispatches callback-style after-auth request handlers.
func (s *Service) DispatchAfterAuthRequest(
	ctx context.Context,
	input pluginhost.AfterAuthInput,
) {
	if input == nil {
		return
	}

	manifests, err := s.scanPluginManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for after-auth dispatch failed: %v", err)
		return
	}

	for _, manifest := range manifests {
		if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetAfterAuthHandlers() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			if err = handler.Handler(ctx, input); err != nil {
				logger.Warningf(ctx, "plugin after-auth handler failed plugin=%s err=%v", manifest.ID, err)
			}
		}
	}
}

func (s *Service) shouldKeepMenu(ctx context.Context, menu *entity.SysMenu) bool {
	if menu == nil {
		return false
	}

	// Translate the internal menu entity into the published contract before invoking
	// plugin filters so plugins never depend on host-only model types.
	descriptor := pluginhost.NewMenuDescriptor(
		menu.Id,
		menu.ParentId,
		menu.Name,
		menu.Path,
		menu.Component,
		menu.Perms,
		menu.MenuKey,
		menu.Type,
		menu.Visible,
		menu.Status,
	)

	manifests, err := s.scanPluginManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for menu filter failed: %v", err)
		return true
	}
	for _, manifest := range manifests {
		if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetMenuFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			visible, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin menu filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !visible {
				return false
			}
		}
	}
	return true
}

// ShouldKeepPermission reports whether a permission should stay effective after plugin filtering.
func (s *Service) ShouldKeepPermission(ctx context.Context, menu *entity.SysMenu) bool {
	if menu == nil {
		return false
	}

	descriptor := pluginhost.NewPermissionDescriptor(
		menu.MenuKey,
		menu.Name,
		menu.Perms,
	)

	manifests, err := s.scanPluginManifests()
	if err != nil {
		logger.Warningf(ctx, "scan plugin manifests for permission filter failed: %v", err)
		return true
	}
	for _, manifest := range manifests {
		if !s.IsEnabled(ctx, manifest.ID) {
			continue
		}
		sourcePlugin, ok := pluginhost.GetSourcePlugin(manifest.ID)
		if !ok {
			continue
		}
		for _, handler := range sourcePlugin.GetPermissionFilters() {
			if handler == nil || handler.Handler == nil {
				continue
			}
			allowed, filterErr := handler.Handler(ctx, descriptor)
			if filterErr != nil {
				logger.Warningf(ctx, "plugin permission filter failed plugin=%s err=%v", manifest.ID, filterErr)
				continue
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}

func (s *Service) buildBackgroundEnabledChecker() pluginhost.PluginEnabledChecker {
	return func(pluginID string) bool {
		return s.IsEnabled(context.Background(), pluginID)
	}
}

func (s *Service) buildPrimaryNodeChecker() pluginhost.PrimaryNodeChecker {
	checker := getPrimaryNodeChecker()
	if checker == nil {
		return nil
	}
	return func() bool {
		return checker()
	}
}
