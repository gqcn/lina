// Package plugin implements plugin manifest discovery, lifecycle orchestration,
// governance metadata synchronization, and host integration for Lina plugins.
package plugin

import (
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
)

// Re-export sub-package types that are referenced by controllers and callers.
type (
	// PluginItem is the display-ready projection of one plugin entry.
	PluginItem = runtime.PluginItem

	// DynamicUploadInput defines input for uploading a runtime WASM package.
	DynamicUploadInput = runtime.DynamicUploadInput

	// DynamicUploadOutput defines output for uploading a runtime WASM package.
	DynamicUploadOutput = runtime.DynamicUploadOutput

	// RuntimeStateListOutput defines output for public runtime state queries.
	RuntimeStateListOutput = runtime.RuntimeStateListOutput

	// ResourceListInput defines input for querying a plugin-owned backend resource.
	ResourceListInput = integration.ResourceListInput

	// ResourceListOutput defines output for querying a plugin-owned backend resource.
	ResourceListOutput = integration.ResourceListOutput

	// RuntimeFrontendAssetOutput contains one resolved frontend asset ready to be served.
	RuntimeFrontendAssetOutput = frontend.RuntimeFrontendAssetOutput

	// DynamicRouteOperLogMetadata stores operation-log metadata for dynamic routes.
	DynamicRouteOperLogMetadata = runtime.DynamicRouteOperLogMetadata

	// PluginDynamicStateItem represents public runtime state of one plugin.
	PluginDynamicStateItem = runtime.PluginDynamicStateItem
)

// GetDynamicRouteOperLogMetadata returns dynamic-route operation-log metadata from the request.
// This package-level function is retained for callers that cannot import the runtime sub-package.
var GetDynamicRouteOperLogMetadata = runtime.GetDynamicRouteOperLogMetadata

// ListOutput defines output for plugin list query.
type ListOutput struct {
	// List contains the filtered plugin list.
	List []*PluginItem
	// Total is the number of returned plugins.
	Total int
}

// ListInput defines input for plugin list query.
type ListInput struct {
	// ID filters by plugin identifier.
	ID string
	// Name filters by plugin display name.
	Name string
	// Type filters by normalized plugin type.
	Type string
	// Status filters by enabled flag.
	Status *int
	// Installed filters by installed flag.
	Installed *int
}

// AuthLoginSucceededInput defines input for auth hook events.
type AuthLoginSucceededInput struct {
	// UserName is the authenticated username.
	UserName string
	// Status is the login status code.
	Status int
	// Ip is the client IP address.
	Ip string
	// ClientType identifies the login client type.
	ClientType string
	// Browser is the detected browser description.
	Browser string
	// Os is the detected operating-system description.
	Os string
	// Message is the audit message delivered to plugins.
	Message string
}

// Service is the plugin system facade. It composes the catalog, lifecycle,
// runtime, integration, frontend, and openapi sub-services and exposes the
// public API used by controllers, middleware, and other host components.
type Service struct {
	// catalogSvc provides manifest discovery, registry, and release governance.
	catalogSvc *catalog.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc *lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc *runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc *integration.Service
	// frontendSvc manages in-memory frontend bundles for dynamic plugins.
	frontendSvc *frontend.Service
	// openapiSvc projects dynamic routes into the host OpenAPI document.
	openapiSvc *openapi.Service
}

// New creates and returns a new plugin Service. An optional Topology may be
// provided for cluster-aware deployments; single-node mode is the default.
func New(topologies ...Topology) *Service {
	var topo Topology = singleNodeTopology{}
	if len(topologies) > 0 && topologies[0] != nil {
		topo = topologies[0]
	}

	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc)
		integrationSvc = integration.New(catalogSvc)
	)

	// Wire cross-package dependencies via setter injection so each sub-package
	// can be constructed independently without circular imports.
	catalogSvc.SetBackendLoader(integrationSvc)
	catalogSvc.SetArtifactParser(runtimeSvc)
	catalogSvc.SetDynamicManifestLoader(runtimeSvc)
	catalogSvc.SetNodeStateSyncer(runtimeSvc)
	catalogSvc.SetMenuSyncer(integrationSvc)
	catalogSvc.SetResourceRefSyncer(integrationSvc)
	catalogSvc.SetReleaseStateSyncer(runtimeSvc)
	catalogSvc.SetHookDispatcher(integrationSvc)

	lifecycleSvc.SetReconciler(runtimeSvc)
	lifecycleSvc.SetTopology(&lifecycleTopologyAdapter{topo})

	integrationSvc.SetBizCtxProvider(&bizCtxAdapter{bizCtxProvider})
	integrationSvc.SetTopologyProvider(&integrationTopologyAdapter{topo})

	runtimeSvc.SetTopology(&runtimeTopologyAdapter{topo})
	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetAfterAuthDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{bizCtxProvider})

	return &Service{
		catalogSvc:     catalogSvc,
		lifecycleSvc:   lifecycleSvc,
		runtimeSvc:     runtimeSvc,
		integrationSvc: integrationSvc,
		frontendSvc:    frontendSvc,
		openapiSvc:     openapiSvc,
	}
}
