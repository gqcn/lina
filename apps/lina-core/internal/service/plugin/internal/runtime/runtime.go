// Package runtime provides the dynamic plugin execution environment: WASM artifact
// parsing, upload handling, background reconciliation, per-node state projection,
// and route dispatch for enabled dynamic plugins.

package runtime

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/pkg/pluginhost"
)

// TopologyProvider abstracts cluster topology information needed by the reconciler.
type TopologyProvider interface {
	// IsClusterModeEnabled reports whether multi-node cluster mode is active.
	IsClusterModeEnabled() bool
	// IsPrimaryNode reports whether this host instance is the designated primary node.
	IsPrimaryNode() bool
	// CurrentNodeID returns the stable host-unique identifier for the current node.
	CurrentNodeID() string
}

// MenuManager abstracts menu-sync operations so the runtime package does not
// directly import the integration package (which depends on runtime).
type MenuManager interface {
	// SyncPluginMenusAndPermissions synchronizes all plugin menus and dynamic route
	// permission entries for the given manifest.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error
	// SyncPluginMenus synchronizes only the declared manifest menus, skipping
	// route-permission entries. Used during rollback to restore a previous menu state.
	SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error
	// DeletePluginMenusByManifest removes all plugin-owned menu rows for the given manifest.
	DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error
}

// HookDispatcher abstracts hook event dispatch so the runtime package does not
// depend on the integration package directly.
type HookDispatcher interface {
	// DispatchPluginHookEvent fires a lifecycle hook event to all registered listeners.
	DispatchPluginHookEvent(
		ctx context.Context,
		event pluginhost.ExtensionPoint,
		values map[string]interface{},
	) error
}

// JwtConfigProvider provides JWT configuration for dynamic route token validation.
type JwtConfigProvider interface {
	// GetJwtSecret returns the JWT signing secret used to validate bearer tokens.
	GetJwtSecret(ctx context.Context) string
}

// UserContextSetter injects authenticated user information into the request context.
type UserContextSetter interface {
	// SetUser populates the context with the resolved token and user identity fields.
	SetUser(ctx context.Context, tokenID string, userID int, username string, status int)
}

// AfterAuthDispatcher fires post-authentication callbacks registered by source plugins.
type AfterAuthDispatcher interface {
	// DispatchAfterAuth invokes all registered after-auth hook handlers.
	DispatchAfterAuth(ctx context.Context, input pluginhost.AfterAuthInput)
}

// PermissionMenuFilter filters button-type permission menus based on plugin enablement.
type PermissionMenuFilter interface {
	// FilterPermissionMenus returns only the menus that pass plugin-level enablement checks.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

// Service coordinates dynamic plugin lifecycle reconciliation, artifact management,
// upload handling, node-state projection, and route dispatch.
type Service struct {
	// catalogSvc provides manifest, registry, and release access.
	catalogSvc *catalog.Service
	// lifecycleSvc provides install/uninstall SQL migration support.
	lifecycleSvc *lifecycle.Service
	// frontendSvc manages in-memory frontend bundles.
	frontendSvc *frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc *openapi.Service
	// topology provides cluster topology information.
	topology TopologyProvider
	// menuMgr handles plugin menu and permission synchronization.
	menuMgr MenuManager
	// hookDispatcher fires lifecycle hook events.
	hookDispatcher HookDispatcher
	// jwtConfig provides the JWT signing secret for route token validation.
	jwtConfig JwtConfigProvider
	// userCtx injects the authenticated user identity into the request context.
	userCtx UserContextSetter
	// afterAuth dispatches post-authentication callbacks to registered source plugins.
	afterAuth AfterAuthDispatcher
	// menuFilter filters button-type permission menus by plugin enablement.
	menuFilter PermissionMenuFilter
}

// New creates a new runtime Service with the given sub-service dependencies.
func New(
	catalogSvc *catalog.Service,
	lifecycleSvc *lifecycle.Service,
	frontendSvc *frontend.Service,
	openapiSvc *openapi.Service,
) *Service {
	return &Service{
		catalogSvc:   catalogSvc,
		lifecycleSvc: lifecycleSvc,
		frontendSvc:  frontendSvc,
		openapiSvc:   openapiSvc,
	}
}

// SetTopology wires the cluster topology provider.
func (s *Service) SetTopology(t TopologyProvider) {
	s.topology = t
}

// SetMenuManager wires the menu synchronization provider.
func (s *Service) SetMenuManager(m MenuManager) {
	s.menuMgr = m
}

// SetHookDispatcher wires the lifecycle hook dispatcher.
func (s *Service) SetHookDispatcher(d HookDispatcher) {
	s.hookDispatcher = d
}

// SetJwtConfigProvider wires the JWT configuration provider for route token validation.
func (s *Service) SetJwtConfigProvider(p JwtConfigProvider) {
	s.jwtConfig = p
}

// SetUserContextSetter wires the user-context injection provider.
func (s *Service) SetUserContextSetter(p UserContextSetter) {
	s.userCtx = p
}

// SetAfterAuthDispatcher wires the post-authentication callback dispatcher.
func (s *Service) SetAfterAuthDispatcher(d AfterAuthDispatcher) {
	s.afterAuth = d
}

// SetPermissionMenuFilter wires the plugin-level permission menu filter.
func (s *Service) SetPermissionMenuFilter(f PermissionMenuFilter) {
	s.menuFilter = f
}

// isClusterModeEnabled is a nil-safe wrapper around the topology provider.
func (s *Service) isClusterModeEnabled() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsClusterModeEnabled()
}

// isPrimaryNode is a nil-safe wrapper around the topology provider.
func (s *Service) isPrimaryNode() bool {
	if s.topology == nil {
		return false
	}
	return s.topology.IsPrimaryNode()
}

// currentNodeID is a nil-safe wrapper around the topology provider.
func (s *Service) currentNodeID() string {
	if s.topology == nil {
		return ""
	}
	return s.topology.CurrentNodeID()
}

// dispatchHookEvent is a nil-safe wrapper for hook event dispatch.
func (s *Service) dispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	if s.hookDispatcher == nil {
		return nil
	}
	return s.hookDispatcher.DispatchPluginHookEvent(ctx, event, values)
}

// syncPluginMenusAndPermissions is a nil-safe wrapper for menu synchronization.
func (s *Service) syncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenusAndPermissions(ctx, manifest)
}

// syncPluginMenus is a nil-safe wrapper for partial menu synchronization (rollback path).
func (s *Service) syncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.SyncPluginMenus(ctx, manifest)
}

// deletePluginMenusByManifest is a nil-safe wrapper for menu deletion.
func (s *Service) deletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuMgr == nil {
		return nil
	}
	return s.menuMgr.DeletePluginMenusByManifest(ctx, manifest)
}

// ensureFrontendBundle delegates to frontendSvc to guarantee an in-memory bundle exists.
func (s *Service) ensureFrontendBundle(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.EnsureBundle(ctx, manifest)
}

// validateFrontendMenuBindings delegates frontend menu binding validation.
func (s *Service) validateFrontendMenuBindings(ctx context.Context, manifest *catalog.Manifest) error {
	if s.frontendSvc == nil {
		return nil
	}
	return s.frontendSvc.ValidateRuntimeFrontendMenuBindings(ctx, manifest)
}

// invalidateFrontendBundle removes all cached frontend bundle entries for a plugin.
func (s *Service) invalidateFrontendBundle(ctx context.Context, pluginID string, reason string) {
	if s.frontendSvc != nil {
		s.frontendSvc.InvalidateBundle(ctx, pluginID, reason)
	}
}
