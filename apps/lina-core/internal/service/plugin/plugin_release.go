// This file synchronizes release-level plugin metadata snapshots into the
// governance tables used by the host management and review workflows.

package plugin

import (
	"context"
	"path"
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

	releaseID := 0
	if existing != nil {
		releaseID = existing.Id
	}
	releaseStatus := s.buildPluginReleaseStatusForManifest(manifest, registry, releaseID)
	// Persist only review-oriented locators and summary snapshots here. Concrete SQL
	// files and frontend source paths are intentionally excluded from table storage.
	data := do.SysPluginRelease{
		PluginId:         manifest.ID,
		ReleaseVersion:   manifest.Version,
		Type:             manifest.Type,
		RuntimeKind:      s.buildPluginDynamicKind(manifest),
		Status:           releaseStatus.String(),
		ManifestPath:     s.buildPluginReleaseManifestPath(manifest),
		PackagePath:      s.buildPluginReleasePackagePathForSync(manifest, existing),
		Checksum:         s.buildPluginRegistryChecksum(manifest),
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

func (s *Service) getPluginReleaseByID(ctx context.Context, releaseID int) (*entity.SysPluginRelease, error) {
	if releaseID <= 0 {
		return nil, nil
	}

	var release *entity.SysPluginRelease
	err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: releaseID}).
		Scan(&release)
	return release, err
}

func (s *Service) getPluginRegistryRelease(ctx context.Context, registry *entity.SysPlugin) (*entity.SysPluginRelease, error) {
	if registry == nil {
		return nil, nil
	}
	if registry.ReleaseId > 0 {
		release, err := s.getPluginReleaseByID(ctx, registry.ReleaseId)
		if err != nil {
			return nil, err
		}
		if release != nil {
			return release, nil
		}
	}
	if strings.TrimSpace(registry.Version) == "" {
		return nil, nil
	}
	return s.getPluginRelease(ctx, registry.PluginId, registry.Version)
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

func (s *Service) buildPluginReleaseStatusForManifest(
	manifest *pluginManifest,
	registry *entity.SysPlugin,
	releaseID int,
) pluginReleaseStatus {
	if manifest == nil || registry == nil {
		return pluginReleaseStatusPrepared
	}
	if normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return s.buildPluginReleaseStatus(registry.Installed, registry.Status)
	}
	if registry.ReleaseId > 0 && releaseID == registry.ReleaseId {
		return s.buildPluginReleaseStatus(registry.Installed, registry.Status)
	}
	if strings.TrimSpace(registry.Version) == strings.TrimSpace(manifest.Version) && registry.ReleaseId <= 0 {
		return s.buildPluginReleaseStatus(registry.Installed, registry.Status)
	}
	return pluginReleaseStatusPrepared
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
		RuntimeKind:               s.buildPluginDynamicKind(manifest),
		RuntimeABIVersion:         s.buildPluginDynamicABIVersion(manifest),
		ManifestDeclared:          s.isPluginManifestDeclared(manifest),
		InstallSQLCount:           s.countPluginSQLAssets(manifest, pluginMigrationDirectionInstall),
		UninstallSQLCount:         s.countPluginSQLAssets(manifest, pluginMigrationDirectionUninstall),
		FrontendPageCount:         s.buildPluginFrontendPageCount(manifest),
		FrontendSlotCount:         s.buildPluginFrontendSlotCount(manifest),
		MenuCount:                 s.buildPluginMenuCount(manifest),
		BackendHookCount:          len(manifest.Hooks),
		ResourceSpecCount:         len(manifest.BackendResources),
		RuntimeFrontendAssetCount: s.buildPluginDynamicFrontendAssetCount(manifest),
		RuntimeSQLAssetCount:      s.buildPluginDynamicSQLAssetCount(manifest),
	}

	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", gerror.Wrap(err, "failed to build plugin manifest snapshot")
	}
	return string(content), nil
}

func (s *Service) buildPluginDynamicKind(manifest *pluginManifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.RuntimeKind
}

func (s *Service) buildPluginDynamicABIVersion(manifest *pluginManifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.ABIVersion
}

func (s *Service) buildPluginDynamicFrontendAssetCount(manifest *pluginManifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.FrontendAssetCount
}

func (s *Service) buildPluginDynamicSQLAssetCount(manifest *pluginManifest) int {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return 0
	}
	return manifest.RuntimeArtifact.SQLAssetCount
}

func (s *Service) countPluginSQLAssets(manifest *pluginManifest, direction pluginMigrationDirection) int {
	assets, err := s.resolvePluginSQLAssets(manifest, direction)
	if err != nil {
		return 0
	}
	return len(assets)
}

func (s *Service) buildPluginPackagePath(manifest *pluginManifest) string {
	if manifest == nil {
		return ""
	}
	if hasSourcePluginEmbeddedFiles(manifest) {
		return "embedded/source-plugins/" + manifest.ID
	}
	if manifest.RuntimeArtifact != nil && strings.TrimSpace(manifest.RuntimeArtifact.Path) != "" {
		normalizedPath := filepath.ToSlash(filepath.Clean(manifest.RuntimeArtifact.Path))
		if marker := "/releases/"; strings.Contains(normalizedPath, marker) {
			return strings.TrimPrefix(normalizedPath[strings.LastIndex(normalizedPath, marker):], "/")
		}
		return filepath.ToSlash(filepath.Base(normalizedPath))
	}
	return filepath.ToSlash(manifest.RootDir)
}

// buildPluginReleasePackagePathForSync keeps archived dynamic-release package
// paths stable. Once a release has been switched to the versioned archive, the
// mutable staging artifact must no longer overwrite that persisted pointer.
func (s *Service) buildPluginReleasePackagePathForSync(
	manifest *pluginManifest,
	existing *entity.SysPluginRelease,
) string {
	if existing != nil {
		existingPackagePath := filepath.ToSlash(strings.TrimSpace(existing.PackagePath))
		if shouldPreserveArchivedPluginReleasePackagePath(manifest, existingPackagePath) {
			return existingPackagePath
		}
	}
	return s.buildPluginPackagePath(manifest)
}

func shouldPreserveArchivedPluginReleasePackagePath(
	manifest *pluginManifest,
	packagePath string,
) bool {
	if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return false
	}
	normalizedPath := filepath.ToSlash(strings.TrimSpace(packagePath))
	if normalizedPath == "" {
		return false
	}
	normalizedPath = strings.TrimPrefix(filepath.Clean("/"+normalizedPath), "/")
	return strings.Contains("/"+normalizedPath, "/releases/")
}

func (s *Service) updatePluginReleaseState(
	ctx context.Context,
	releaseID int,
	status pluginReleaseStatus,
	packagePath string,
) error {
	if releaseID <= 0 {
		return nil
	}

	data := do.SysPluginRelease{
		Status: status.String(),
	}
	if strings.TrimSpace(packagePath) != "" {
		data.PackagePath = filepath.ToSlash(strings.TrimSpace(packagePath))
	}

	_, err := dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: releaseID}).
		Data(data).
		Update()
	return err
}

func (s *Service) buildPluginReleaseManifestPath(manifest *pluginManifest) string {
	if manifest == nil || normalizePluginType(manifest.Type) == pluginTypeDynamic {
		return ""
	}
	if hasSourcePluginEmbeddedFiles(manifest) {
		return path.Clean(strings.ReplaceAll(manifest.ManifestPath, "\\", "/"))
	}
	return filepath.ToSlash(filepath.Base(manifest.ManifestPath))
}

func (s *Service) isPluginManifestDeclared(manifest *pluginManifest) bool {
	if manifest == nil {
		return false
	}
	if strings.TrimSpace(manifest.ManifestPath) != "" {
		return true
	}
	return manifest.RuntimeArtifact != nil && manifest.RuntimeArtifact.Manifest != nil
}

func (s *Service) buildPluginFrontendPageCount(manifest *pluginManifest) int {
	if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeSource {
		return 0
	}
	return len(s.listPluginFrontendPagePaths(manifest))
}

func (s *Service) buildPluginFrontendSlotCount(manifest *pluginManifest) int {
	if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeSource {
		return 0
	}
	return len(s.listPluginFrontendSlotPaths(manifest))
}

func (s *Service) buildPluginMenuCount(manifest *pluginManifest) int {
	if manifest == nil {
		return 0
	}
	return len(manifest.Menus)
}
