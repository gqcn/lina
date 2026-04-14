// Package catalog provides plugin manifest discovery, registry management,
// release tracking, and governance queries for the Lina plugin system.
package catalog

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginhost"
)

// ConfigProvider abstracts the configuration dependency needed for manifest scanning.
type ConfigProvider interface {
	// GetPluginDynamicStoragePath returns the filesystem path where runtime wasm
	// artifacts are stored.
	GetPluginDynamicStoragePath(ctx context.Context) string
}

// BackendConfigLoader loads plugin backend hook/resource declarations into a manifest.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle (integration → catalog → integration).
type BackendConfigLoader interface {
	// LoadPluginBackendConfig populates Hooks and BackendResources on the given manifest.
	LoadPluginBackendConfig(manifest *Manifest) error
}

// ArtifactParser parses a runtime WASM artifact file and extracts its embedded sections.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle (runtime → catalog → runtime).
type ArtifactParser interface {
	// ParseRuntimeWasmArtifact reads and validates the WASM file at filePath.
	ParseRuntimeWasmArtifact(filePath string) (*ArtifactSpec, error)
	// ParseRuntimeWasmArtifactContent parses a WASM artifact from an in-memory byte slice.
	ParseRuntimeWasmArtifactContent(filePath string, content []byte) (*ArtifactSpec, error)
	// ValidateRuntimeArtifact validates a dynamic plugin source-tree artifact against manifest.
	ValidateRuntimeArtifact(manifest *Manifest, rootDir string) error
}

// DynamicManifestLoader loads the currently active manifest for an installed dynamic plugin.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle.
type DynamicManifestLoader interface {
	// LoadActiveDynamicPluginManifest returns the manifest backed by the active archived release.
	LoadActiveDynamicPluginManifest(ctx context.Context, registry *entity.SysPlugin) (*Manifest, error)
}

// NodeStateSyncer synchronizes node-level plugin state records.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle (runtime → catalog → runtime).
type NodeStateSyncer interface {
	// SyncPluginNodeState upserts the node state record for a plugin lifecycle event.
	SyncPluginNodeState(ctx context.Context, pluginID, version string, installed, enabled int, message string) error
	// GetPluginNodeState returns the current node state record for one plugin on one node.
	GetPluginNodeState(ctx context.Context, pluginID, nodeID string) (*entity.SysPluginNodeState, error)
	// CurrentNodeID returns the cluster node identifier for the running host.
	CurrentNodeID() string
}

// MenuSyncer synchronizes plugin-declared menus into the host menu table.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle (integration → catalog → integration).
type MenuSyncer interface {
	// SyncPluginMenusAndPermissions reconciles manifest menus into sys_menu and admin role.
	SyncPluginMenusAndPermissions(ctx context.Context, manifest *Manifest) error
}

// ResourceRefSyncer synchronizes plugin resource reference records.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle.
type ResourceRefSyncer interface {
	// SyncPluginResourceReferences persists resource reference rows for governance review.
	SyncPluginResourceReferences(ctx context.Context, manifest *Manifest) error
}

// ReleaseStateSyncer synchronizes the active runtime state of a plugin release.
// This interface is implemented by the runtime sub-package and injected after
// construction to avoid an import cycle.
type ReleaseStateSyncer interface {
	// SyncPluginReleaseRuntimeState updates the active release row to reflect registry state.
	SyncPluginReleaseRuntimeState(ctx context.Context, registry *entity.SysPlugin) error
}

// HookDispatcher dispatches plugin lifecycle events to registered hook handlers.
// This interface is implemented by the integration sub-package and injected after
// construction to avoid an import cycle.
type HookDispatcher interface {
	// DispatchPluginHookEvent fires a lifecycle hook event with the given payload.
	DispatchPluginHookEvent(ctx context.Context, event pluginhost.ExtensionPoint, values map[string]interface{}) error
}

// Service provides plugin manifest discovery, registry management,
// release tracking, and governance queries.
type Service struct {
	// configSvc provides plugin configuration values.
	configSvc ConfigProvider
	// backendLoader loads backend hook/resource declarations into manifests.
	// Set via SetBackendLoader after construction to avoid import cycles.
	backendLoader BackendConfigLoader
	// artifactParser reads and validates WASM artifact files.
	// Set via SetArtifactParser after construction to avoid import cycles.
	artifactParser ArtifactParser
	// dynamicManifestLoader loads the active release manifest for dynamic plugins.
	// Set via SetDynamicManifestLoader after construction to avoid import cycles.
	dynamicManifestLoader DynamicManifestLoader
	// nodeStateSyncer syncs node state records for lifecycle events.
	// Set via SetNodeStateSyncer after construction to avoid import cycles.
	nodeStateSyncer NodeStateSyncer
	// menuSyncer syncs plugin menus into the host menu table.
	// Set via SetMenuSyncer after construction to avoid import cycles.
	menuSyncer MenuSyncer
	// resourceRefSyncer syncs plugin resource reference records.
	// Set via SetResourceRefSyncer after construction to avoid import cycles.
	resourceRefSyncer ResourceRefSyncer
	// releaseStateSyncer syncs the active runtime state of a plugin release.
	// Set via SetReleaseStateSyncer after construction to avoid import cycles.
	releaseStateSyncer ReleaseStateSyncer
	// hookDispatcher dispatches lifecycle hook events to registered handlers.
	// Set via SetHookDispatcher after construction to avoid import cycles.
	hookDispatcher HookDispatcher
}

// New creates a new catalog Service with the given configuration provider.
// Call the Set* methods after all sub-services are constructed to wire
// the cross-package dependencies.
func New(configSvc ConfigProvider) *Service {
	return &Service{configSvc: configSvc}
}

// SetBackendLoader wires the integration package's backend config loader.
func (s *Service) SetBackendLoader(loader BackendConfigLoader) {
	s.backendLoader = loader
}

// SetArtifactParser wires the runtime package's WASM artifact parser.
func (s *Service) SetArtifactParser(parser ArtifactParser) {
	s.artifactParser = parser
}

// SetDynamicManifestLoader wires the runtime package's active manifest loader.
func (s *Service) SetDynamicManifestLoader(loader DynamicManifestLoader) {
	s.dynamicManifestLoader = loader
}

// SetNodeStateSyncer wires the runtime package's node state syncer.
func (s *Service) SetNodeStateSyncer(syncer NodeStateSyncer) {
	s.nodeStateSyncer = syncer
}

// SetMenuSyncer wires the integration package's menu syncer.
func (s *Service) SetMenuSyncer(syncer MenuSyncer) {
	s.menuSyncer = syncer
}

// SetResourceRefSyncer wires the integration package's resource reference syncer.
func (s *Service) SetResourceRefSyncer(syncer ResourceRefSyncer) {
	s.resourceRefSyncer = syncer
}

// SetReleaseStateSyncer wires the runtime package's release state syncer.
func (s *Service) SetReleaseStateSyncer(syncer ReleaseStateSyncer) {
	s.releaseStateSyncer = syncer
}

// SetHookDispatcher wires the integration package's hook event dispatcher.
func (s *Service) SetHookDispatcher(dispatcher HookDispatcher) {
	s.hookDispatcher = dispatcher
}
