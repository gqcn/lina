// This file exposes host integration methods on the root plugin facade.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// RegisterHTTPRoutes registers callback-contributed HTTP routes for source plugins.
func (s *Service) RegisterHTTPRoutes(
	ctx context.Context,
	pluginGroup *ghttp.RouterGroup,
	middlewares pluginhost.RouteMiddlewares,
) error {
	return s.integrationSvc.RegisterHTTPRoutes(ctx, pluginGroup, middlewares)
}

// RegisterCrons registers callback-contributed cron jobs for source plugins.
func (s *Service) RegisterCrons(ctx context.Context) error {
	return s.integrationSvc.RegisterCrons(ctx)
}

// DispatchAfterAuthRequest dispatches callback-style after-auth request handlers.
func (s *Service) DispatchAfterAuthRequest(ctx context.Context, input pluginhost.AfterAuthInput) {
	s.integrationSvc.DispatchAfterAuth(ctx, input)
}

// DispatchHookEvent dispatches one named hook event to all enabled plugins.
func (s *Service) DispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	return s.integrationSvc.DispatchPluginHookEvent(ctx, event, values)
}

// FilterMenus filters disabled plugin menus from the given menu list.
func (s *Service) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return s.integrationSvc.FilterMenus(ctx, menus)
}

// FilterPermissionMenus filters permission menus based on plugin enablement.
func (s *Service) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return s.integrationSvc.FilterPermissionMenus(ctx, menus)
}

// ListResourceRecords queries plugin-owned backend resource rows.
func (s *Service) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	return s.integrationSvc.ListResourceRecords(ctx, in)
}
