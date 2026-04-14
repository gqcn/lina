// This file exposes runtime and dynamic-route facade methods.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
)

// StartRuntimeReconciler starts the background reconciler loop for dynamic plugins.
func (s *Service) StartRuntimeReconciler(ctx context.Context) {
	s.runtimeSvc.StartRuntimeReconciler(ctx)
}

// ReconcileRuntimePlugins runs one reconciliation pass for all dynamic plugins.
func (s *Service) ReconcileRuntimePlugins(ctx context.Context) error {
	return s.runtimeSvc.ReconcileRuntimePlugins(ctx)
}

// ListRuntimeStates returns public plugin runtime states for shell slot rendering.
func (s *Service) ListRuntimeStates(ctx context.Context) (*RuntimeStateListOutput, error) {
	return s.runtimeSvc.ListRuntimeStates(ctx)
}

// UploadDynamicPackage validates and stores a runtime WASM package.
func (s *Service) UploadDynamicPackage(ctx context.Context, in *DynamicUploadInput) (*DynamicUploadOutput, error) {
	return s.runtimeSvc.UploadDynamicPackage(ctx, in)
}

// PrepareDynamicRouteMiddleware prepares dynamic route state before the main handler.
func (s *Service) PrepareDynamicRouteMiddleware(r *ghttp.Request) {
	s.runtimeSvc.PrepareDynamicRouteMiddleware(r)
}

// AuthenticateDynamicRouteMiddleware authenticates JWT tokens for dynamic routes.
func (s *Service) AuthenticateDynamicRouteMiddleware(r *ghttp.Request) {
	s.runtimeSvc.AuthenticateDynamicRouteMiddleware(r)
}

// RegisterDynamicRouteDispatcher binds the dynamic route catch-all handler to the group.
func (s *Service) RegisterDynamicRouteDispatcher(group *ghttp.RouterGroup) {
	s.runtimeSvc.RegisterDynamicRouteDispatcher(group)
}
