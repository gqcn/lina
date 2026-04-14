// Package lifecycle implements dynamic plugin install, uninstall, and reconcile
// lifecycle flows together with helpers for resolving runtime plugin resources.
package lifecycle

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
)

// ReconcileProvider abstracts the runtime reconciler so lifecycle can trigger
// plugin convergence without importing the runtime sub-package directly.
type ReconcileProvider interface {
	// ReconcileDynamicPluginRequest submits a desired state transition to the reconciler.
	ReconcileDynamicPluginRequest(ctx context.Context, pluginID string, desiredState string) error
	// ShouldRefreshInstalledDynamicRelease reports whether the installed release is stale.
	ShouldRefreshInstalledDynamicRelease(ctx context.Context, registry interface{}, manifest *catalog.Manifest) bool
	// EnsureRuntimeArtifactAvailable ensures the WASM artifact is present for lifecycle actions.
	EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error
}

// TopologyProvider abstracts the cluster topology status needed by lifecycle flows.
type TopologyProvider interface {
	// IsPrimaryNode reports whether this host instance is the primary cluster node.
	IsPrimaryNode() bool
}

// Service provides install, uninstall, and migration lifecycle orchestration for dynamic plugins.
type Service struct {
	// catalogSvc provides manifest discovery and registry access.
	catalogSvc *catalog.Service
	// reconciler triggers runtime convergence for desired state transitions.
	reconciler ReconcileProvider
	// topology provides cluster topology information.
	topology TopologyProvider
}

// New creates a new lifecycle Service with the given catalog service.
// Call SetReconciler and SetTopology after construction to wire runtime dependencies.
func New(catalogSvc *catalog.Service) *Service {
	return &Service{catalogSvc: catalogSvc}
}

// SetReconciler wires the runtime package's reconcile provider.
func (s *Service) SetReconciler(r ReconcileProvider) {
	s.reconciler = r
}

// SetTopology wires the cluster topology provider.
func (s *Service) SetTopology(t TopologyProvider) {
	s.topology = t
}

// Install executes the install lifecycle for a discovered dynamic plugin.
// Repeated installs are treated as idempotent unless the same version needs a refresh.
func (s *Service) Install(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return gerror.New("源码插件随宿主编译集成，不支持安装")
	}
	if s.reconciler != nil {
		if err = s.reconciler.EnsureRuntimeArtifactAvailable(manifest, "安装"); err != nil {
			return err
		}
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, manifest)
	if err != nil {
		return err
	}
	if registry.Installed == catalog.InstalledYes {
		compareResult, compareErr := catalog.CompareSemanticVersions(manifest.Version, registry.Version)
		if compareErr != nil {
			return compareErr
		}
		if compareResult < 0 {
			return gerror.New("不支持回退到更低版本，请使用宿主自动回滚结果或重新上传更高版本")
		}
		if compareResult == 0 {
			if s.reconciler != nil && !s.reconciler.ShouldRefreshInstalledDynamicRelease(ctx, registry, manifest) {
				return nil
			}
		}
	}

	desiredState := catalog.HostStateInstalled.String()
	if registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled {
		desiredState = catalog.HostStateEnabled.String()
	}
	if s.reconciler != nil {
		if err = s.reconciler.ReconcileDynamicPluginRequest(ctx, pluginID, desiredState); err != nil {
			return err
		}
	}
	return nil
}

// Uninstall executes the uninstall lifecycle for an installed dynamic plugin.
func (s *Service) Uninstall(ctx context.Context, pluginID string) error {
	manifest, err := s.catalogSvc.GetDesiredManifest(pluginID)
	if err != nil {
		return err
	}
	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		return gerror.New("源码插件随宿主编译集成，不支持卸载")
	}

	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != catalog.InstalledYes {
		return nil
	}
	if s.reconciler != nil {
		return s.reconciler.ReconcileDynamicPluginRequest(ctx, pluginID, catalog.HostStateUninstalled.String())
	}
	return nil
}
