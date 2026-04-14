// Package integration bridges pluginhost callback registrations and declared plugin
// configurations into the host route, menu, permission, and lifecycle integration flows.

package integration

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginhost"
)

// BizCtxProvider abstracts the business context dependency for data-scope queries.
type BizCtxProvider interface {
	// GetUserId returns the user ID stored in the current request business context.
	GetUserId(ctx context.Context) int
}

// TopologyProvider abstracts cluster topology for primary-node routing decisions.
type TopologyProvider interface {
	// IsPrimaryNode reports whether this host instance is the designated primary node.
	IsPrimaryNode() bool
}

// filterRuntime holds a snapshot of which plugins are currently enabled for use
// by menu and permission filters within a single request.
type filterRuntime struct {
	manifests   []*catalog.Manifest
	enabledByID map[string]bool
}

// isEnabled reports whether the plugin with the given ID is currently enabled.
func (r *filterRuntime) isEnabled(pluginID string) bool {
	if r == nil {
		return false
	}
	return r.enabledByID[strings.TrimSpace(pluginID)]
}

// Service bridges plugin callbacks and declarations into host integration points.
type Service struct {
	// catalogSvc provides manifest discovery, registry queries, and release access.
	catalogSvc *catalog.Service
	// bizCtxSvc provides the current user ID for data-scope queries.
	bizCtxSvc BizCtxProvider
	// topology provides cluster topology for primary-node route checks.
	topology TopologyProvider
}

// New creates a new integration Service backed by the given catalog service.
func New(catalogSvc *catalog.Service) *Service {
	return &Service{catalogSvc: catalogSvc}
}

// SetBizCtxProvider wires the business context dependency for data-scope queries.
func (s *Service) SetBizCtxProvider(p BizCtxProvider) {
	s.bizCtxSvc = p
}

// SetTopologyProvider wires the cluster topology provider.
func (s *Service) SetTopologyProvider(t TopologyProvider) {
	s.topology = t
}

// IsEnabled reports whether the plugin with the given ID is currently installed and enabled.
func (s *Service) IsEnabled(ctx context.Context, pluginID string) bool {
	registry, err := s.catalogSvc.GetRegistry(ctx, pluginID)
	if err != nil || registry == nil {
		return false
	}
	return registry.Installed == catalog.InstalledYes && registry.Status == catalog.StatusEnabled
}

// buildFilterRuntime builds a filter runtime by scanning all manifests and loading
// the current enablement status for each discovered plugin.
func (s *Service) buildFilterRuntime(ctx context.Context) (*filterRuntime, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}
	return s.buildFilterRuntimeFromManifests(ctx, manifests)
}

// buildFilterRuntimeFromManifests builds a filter runtime for the given manifest list.
func (s *Service) buildFilterRuntimeFromManifests(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (*filterRuntime, error) {
	enabledByID, err := s.buildEnabledPluginMap(ctx, manifests)
	if err != nil {
		return nil, err
	}
	return &filterRuntime{
		manifests:   manifests,
		enabledByID: enabledByID,
	}, nil
}

// buildEnabledPluginMap queries the registry table for the installed/enabled state
// of each plugin in the manifest list.
func (s *Service) buildEnabledPluginMap(
	ctx context.Context,
	manifests []*catalog.Manifest,
) (map[string]bool, error) {
	enabledByID := make(map[string]bool, len(manifests))
	pluginIDs := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		if manifest == nil {
			continue
		}
		pluginID := strings.TrimSpace(manifest.ID)
		if pluginID == "" {
			continue
		}
		if _, ok := enabledByID[pluginID]; ok {
			continue
		}
		enabledByID[pluginID] = false
		pluginIDs = append(pluginIDs, pluginID)
	}
	if len(pluginIDs) == 0 {
		return enabledByID, nil
	}

	var registries []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		WhereIn(dao.SysPlugin.Columns().PluginId, pluginIDs).
		Scan(&registries)
	if err != nil {
		return nil, err
	}

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		pluginID := strings.TrimSpace(registry.PluginId)
		enabledByID[pluginID] = registry.Installed == catalog.InstalledYes &&
			registry.Status == catalog.StatusEnabled
	}
	return enabledByID, nil
}

// buildBackgroundEnabledChecker returns a PluginEnabledChecker for use in source plugin
// route and cron registrars that need to guard runtime access.
func (s *Service) buildBackgroundEnabledChecker() pluginhost.PluginEnabledChecker {
	return func(pluginID string) bool {
		return s.IsEnabled(context.Background(), pluginID)
	}
}

// buildPrimaryNodeChecker returns a PrimaryNodeChecker for use in source plugin cron registrars.
func (s *Service) buildPrimaryNodeChecker() pluginhost.PrimaryNodeChecker {
	return func() bool {
		if s.topology == nil {
			return false
		}
		return s.topology.IsPrimaryNode()
	}
}
