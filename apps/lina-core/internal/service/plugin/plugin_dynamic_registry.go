package plugin

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
)

func (s *Service) listRuntimeRegistries(ctx context.Context) ([]*entity.SysPlugin, error) {
	var list []*entity.SysPlugin
	err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{Type: pluginTypeDynamic.String()}).
		OrderAsc(dao.SysPlugin.Columns().PluginId).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) buildPluginItem(manifest *pluginManifest, registry *entity.SysPlugin) *PluginItem {
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
		StatusKey:   s.buildPluginStatusKey(id),
		UpdatedAt:   updatedAt,
	}
}

func (s *Service) hasRuntimeArtifactStorageFile(ctx context.Context, pluginID string) (bool, string, error) {
	storageDir, err := s.resolveRuntimePluginStorageDir(ctx)
	if err != nil {
		return false, "", err
	}

	targetPath := filepath.Join(storageDir, buildPluginDynamicArtifactFileName(pluginID))
	if gfile.Exists(targetPath) {
		return true, targetPath, nil
	}

	conflictPath, err := s.findDuplicateRuntimeArtifactPath(storageDir, pluginID, targetPath)
	if err != nil {
		return false, "", err
	}
	if conflictPath != "" {
		return true, conflictPath, nil
	}
	return false, targetPath, nil
}

func (s *Service) reconcileRuntimeRegistryArtifactState(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPlugin, error) {
	if registry == nil || normalizePluginType(registry.Type) != pluginTypeDynamic {
		return registry, nil
	}
	if strings.TrimSpace(registry.PluginId) == "" {
		return registry, nil
	}

	exists, _, err := s.hasRuntimeArtifactStorageFile(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if exists {
		return registry, nil
	}
	if registry.Installed != pluginInstalledYes && registry.Status != pluginStatusEnabled {
		return registry, nil
	}

	data := do.SysPlugin{
		Installed:    pluginInstalledNo,
		Status:       pluginStatusDisabled,
		DesiredState: pluginHostStateUninstalled.String(),
		CurrentState: pluginHostStateUninstalled.String(),
		ReleaseId:    0,
		Generation:   nextPluginGeneration(registry),
		DisabledAt:   gtime.Now(),
	}
	if _, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update(); err != nil {
		return nil, err
	}

	s.invalidateRuntimeFrontendBundle(ctx, registry.PluginId, "runtime_artifact_missing")

	updated, err := s.getPluginRegistry(ctx, registry.PluginId)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	if err = s.syncPluginReleaseRuntimeState(ctx, updated); err != nil {
		return nil, err
	}
	if err = s.syncPluginNodeState(
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

func (s *Service) syncPluginReleaseRuntimeState(ctx context.Context, registry *entity.SysPlugin) error {
	if registry == nil || normalizePluginType(registry.Type) != pluginTypeDynamic {
		return nil
	}

	release, err := s.getPluginRelease(ctx, registry.PluginId, registry.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return nil
	}

	status := s.buildPluginReleaseStatus(registry.Installed, registry.Status)
	_, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: release.Id}).
		Data(do.SysPluginRelease{Status: status.String()}).
		Update()
	return err
}

func sortPluginItems(items []*PluginItem) {
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
