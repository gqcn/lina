// This file provides registry-level helpers used by the reconciler and dynamic
// state projections: listing runtime registries, checking artifact file existence,
// and reconciling registry rows when artifacts are missing from storage.

package runtime

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// PluginItem is a flattened, display-ready projection of one plugin entry combining
// manifest fields with the live registry row for management API responses.
type PluginItem struct {
	// Id is the stable plugin identifier.
	Id string
	// Name is the human-readable display name.
	Name string
	// Version is the currently active version string.
	Version string
	// Type is the normalized plugin type (source or dynamic).
	Type string
	// Description is the short plugin description.
	Description string
	// Installed reports whether the plugin has been installed.
	Installed int
	// InstalledAt is the ISO timestamp of first installation.
	InstalledAt string
	// Enabled reports whether the plugin is currently enabled.
	Enabled int
	// StatusKey is the host config key used by the public shell.
	StatusKey string
	// UpdatedAt is the ISO timestamp of the last registry update.
	UpdatedAt string
}

// listRuntimeRegistries returns all dynamic-type plugin registry rows.
func (s *Service) listRuntimeRegistries(ctx context.Context) ([]*entity.SysPlugin, error) {
	var list []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{Type: catalog.TypeDynamic.String()}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// buildPluginItem returns a PluginItem projection combining manifest and registry data.
func (s *Service) buildPluginItem(manifest *catalog.Manifest, registry *entity.SysPlugin) *PluginItem {
	if manifest == nil && registry == nil {
		return nil
	}

	var (
		id          string
		name        string
		version     string
		pluginType  string
		description string
		installed   int
		enabled     int
		installedAt string
		updatedAt   string
	)

	if manifest != nil {
		id = manifest.ID
		name = manifest.Name
		version = manifest.Version
		pluginType = manifest.Type
		description = manifest.Description
	}
	if registry != nil {
		if registry.PluginId != "" {
			id = registry.PluginId
		}
		if registry.Name != "" {
			name = registry.Name
		}
		if registry.Version != "" {
			version = registry.Version
		}
		if registry.Type != "" {
			pluginType = registry.Type
		}
		if registry.Remark != "" {
			description = registry.Remark
		}
		installed = registry.Installed
		enabled = registry.Status
		if registry.InstalledAt != nil {
			installedAt = registry.InstalledAt.String()
		}
		if registry.UpdatedAt != nil {
			updatedAt = registry.UpdatedAt.String()
		}
	}

	return &PluginItem{
		Id:          id,
		Name:        name,
		Version:     version,
		Type:        pluginType,
		Description: description,
		Installed:   installed,
		InstalledAt: installedAt,
		Enabled:     enabled,
		StatusKey:   s.catalogSvc.BuildPluginStatusKey(id),
		UpdatedAt:   updatedAt,
	}
}

// hasArtifactStorageFile reports whether the runtime artifact for pluginID exists
// in the configured storage directory.
func (s *Service) hasArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error) {
	storageDir, err := s.catalogSvc.RuntimeStorageDir(ctx)
	if err != nil {
		return false, "", err
	}

	targetPath := filepath.Join(storageDir, buildArtifactFileName(pluginID))
	if gfile.Exists(targetPath) {
		return true, targetPath, nil
	}

	conflictPath, err := s.findDuplicateArtifactPath(storageDir, pluginID, targetPath)
	if err != nil {
		return false, "", err
	}
	if conflictPath != "" {
		return true, conflictPath, nil
	}
	return false, targetPath, nil
}

// HasArtifactStorageFile is the exported form of hasArtifactStorageFile for cross-package access.
func (s *Service) HasArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error) {
	return s.hasArtifactStorageFile(ctx, pluginID)
}

// reconcileRegistryArtifactState resets a dynamic plugin registry row to
// uninstalled when its runtime artifact file can no longer be found on disk.
func (s *Service) reconcileRegistryArtifactState(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPlugin, error) {
	if registry == nil || catalog.NormalizeType(registry.Type) != catalog.TypeDynamic {
		return registry, nil
	}
	if strings.TrimSpace(registry.PluginId) == "" {
		return registry, nil
	}

	exists, _, err := s.hasArtifactStorageFile(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if exists {
		return registry, nil
	}
	if registry.Installed != catalog.InstalledYes && registry.Status != catalog.StatusEnabled {
		return registry, nil
	}

	data := do.SysPlugin{
		Installed:    catalog.InstalledNo,
		Status:       catalog.StatusDisabled,
		DesiredState: catalog.HostStateUninstalled.String(),
		CurrentState: catalog.HostStateUninstalled.String(),
		ReleaseId:    0,
		Generation:   catalog.NextGeneration(registry),
		DisabledAt:   gtime.Now(),
	}
	if _, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update(); err != nil {
		return nil, err
	}

	s.invalidateFrontendBundle(ctx, registry.PluginId, "runtime_artifact_missing")

	updated, err := s.catalogSvc.GetRegistry(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	if err = s.SyncPluginReleaseRuntimeState(ctx, updated); err != nil {
		return nil, err
	}
	if err = s.SyncPluginNodeState(
		ctx,
		updated.PluginId,
		updated.Version,
		updated.Installed,
		updated.Status,
		"Runtime plugin artifact missing from storage path; host registry reconciled to uninstalled.",
	); err != nil {
		return nil, err
	}
	return updated, nil
}

// SortPluginItems sorts a PluginItem slice by plugin ID ascending.
func SortPluginItems(items []*PluginItem) {
	sort.Slice(items, func(i int, j int) bool {
		if items[i] == nil {
			return false
		}
		if items[j] == nil {
			return true
		}
		return items[i].Id < items[j].Id
	})
}
