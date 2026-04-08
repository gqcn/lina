// This file synchronizes release-level plugin metadata snapshots into the
// governance tables used by the host management and review workflows.

package plugin

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// syncPluginMetadata persists the review-oriented metadata snapshot after a manifest or lifecycle change.
func (s *Service) syncPluginMetadata(ctx context.Context, manifest *pluginManifest, registry *entity.SysPlugin, message string) error {
	if manifest == nil || registry == nil {
		return nil
	}
	if err := s.syncPluginReleaseMetadata(ctx, manifest, registry); err != nil {
		return err
	}
	if err := s.syncPluginResourceReferences(ctx, manifest); err != nil {
		return err
	}
	return s.syncPluginNodeState(ctx, registry.PluginId, registry.Version, registry.Installed, registry.Status, message)
}

// syncPluginReleaseMetadata upserts the current manifest snapshot into sys_plugin_release.
func (s *Service) syncPluginReleaseMetadata(ctx context.Context, manifest *pluginManifest, registry *entity.SysPlugin) error {
	if manifest == nil || registry == nil {
		return nil
	}

	snapshot, err := s.buildPluginManifestSnapshot(manifest)
	if err != nil {
		return err
	}

	existing, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}

	releaseStatus := s.buildPluginReleaseStatus(registry.Installed, registry.Status)
	// Persist only review-oriented locators and summary snapshots here. Concrete SQL
	// files and frontend source paths are intentionally excluded from table storage.
	data := do.SysPluginRelease{
		PluginId:         manifest.ID,
		ReleaseVersion:   manifest.Version,
		Type:             manifest.Type,
		Status:           releaseStatus.String(),
		ManifestPath:     filepath.ToSlash(filepath.Base(manifest.ManifestPath)),
		PackagePath:      filepath.ToSlash(manifest.RootDir),
		Checksum:         registry.Checksum,
		ManifestSnapshot: snapshot,
	}

	if existing == nil {
		_, err = dao.SysPluginRelease.Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: existing.Id}).
		Data(data).
		Update()
	return err
}

func (s *Service) getPluginRelease(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error) {
	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{
			PluginId:       pluginID,
			ReleaseVersion: version,
		}).
		Scan(&release)
	return release, err
}

func (s *Service) buildPluginReleaseStatus(installed int, enabled int) pluginReleaseStatus {
	if installed != pluginInstalledYes {
		return pluginReleaseStatusUninstalled
	}
	if enabled == pluginStatusEnabled {
		return pluginReleaseStatusActive
	}
	return pluginReleaseStatusInstalled
}

func (s *Service) buildPluginManifestSnapshot(manifest *pluginManifest) (string, error) {
	if manifest == nil {
		return "", gerror.New("plugin manifest cannot be nil")
	}

	snapshot := &pluginManifestSnapshot{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Type:        manifest.Type,
		Description: manifest.Description,
		Author:      manifest.Author,
		Homepage:    manifest.Homepage,
		License:     manifest.License,
		// Record whether the manifest exists without embedding an environment-specific
		// file path into the persisted YAML snapshot.
		ManifestDeclared:  strings.TrimSpace(manifest.ManifestPath) != "",
		InstallSQLCount:   len(s.discoverPluginSQLPaths(manifest.RootDir, false)),
		UninstallSQLCount: len(s.discoverPluginSQLPaths(manifest.RootDir, true)),
		FrontendPageCount: len(s.discoverPluginPagePaths(manifest.RootDir)),
		FrontendSlotCount: len(s.discoverPluginSlotPaths(manifest.RootDir)),
		BackendHookCount:  len(manifest.Hooks),
		ResourceSpecCount: len(manifest.BackendResources),
	}

	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", gerror.Wrap(err, "failed to build plugin manifest snapshot")
	}
	return string(content), nil
}
