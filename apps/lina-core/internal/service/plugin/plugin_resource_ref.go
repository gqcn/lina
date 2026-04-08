// This file synchronizes abstract plugin resource descriptors into
// sys_plugin_resource_ref for framework-agnostic governance review.

package plugin

import (
	"context"
	"fmt"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// syncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the current abstract resource summary.
func (s *Service) syncPluginResourceReferences(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}

	release, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return nil
	}

	existingRefs, err := s.listPluginResourceRefs(ctx, manifest.ID, release.Id)
	if err != nil {
		return err
	}

	existingMap := make(map[string]*entity.SysPluginResourceRef, len(existingRefs))
	for _, item := range existingRefs {
		if item == nil {
			continue
		}
		existingMap[s.buildPluginResourceIdentity(item.ResourceType, item.ResourceKey)] = item
	}

	seen := make(map[string]struct{})
	for _, descriptor := range s.buildPluginResourceRefDescriptors(manifest) {
		identity := s.buildPluginResourceIdentity(descriptor.Kind.String(), descriptor.Key)
		seen[identity] = struct{}{}

		if existing, ok := existingMap[identity]; ok {
			// Only update abstract ownership and review remarks. Concrete file paths are
			// deliberately excluded so the schema stays framework-agnostic.
			_, err = dao.SysPluginResourceRef.Ctx(ctx).
				Where(do.SysPluginResourceRef{Id: existing.Id}).
				Data(do.SysPluginResourceRef{
					OwnerType: descriptor.OwnerType.String(),
					OwnerKey:  descriptor.OwnerKey,
					Remark:    descriptor.Remark,
				}).
				Update()
			if err != nil {
				return err
			}
			continue
		}

		// Persist stable resource identities that describe what the host discovered,
		// not where each file lives inside a framework-specific directory tree.
		_, err = dao.SysPluginResourceRef.Ctx(ctx).Data(do.SysPluginResourceRef{
			PluginId:     manifest.ID,
			ReleaseId:    release.Id,
			ResourceType: descriptor.Kind.String(),
			ResourceKey:  descriptor.Key,
			ResourcePath: "",
			OwnerType:    descriptor.OwnerType.String(),
			OwnerKey:     descriptor.OwnerKey,
			Remark:       descriptor.Remark,
		}).Insert()
		if err != nil {
			return err
		}
	}

	for _, item := range existingRefs {
		if item == nil {
			continue
		}
		identity := s.buildPluginResourceIdentity(item.ResourceType, item.ResourceKey)
		if _, ok := seen[identity]; ok {
			continue
		}
		if _, err = dao.SysPluginResourceRef.Ctx(ctx).
			Where(do.SysPluginResourceRef{Id: item.Id}).
			Delete(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) listPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error) {
	items := make([]*entity.SysPluginResourceRef, 0)
	err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{
			PluginId:  pluginID,
			ReleaseId: releaseID,
		}).
		Scan(&items)
	return items, err
}

// buildPluginResourceRefDescriptors converts concrete discovery results into framework-agnostic review records.
func (s *Service) buildPluginResourceRefDescriptors(manifest *pluginManifest) []*pluginResourceRefDescriptor {
	if manifest == nil {
		return []*pluginResourceRefDescriptor{}
	}

	// The discovery layer still inspects concrete directories so the host can validate
	// the plugin bundle, but the persisted descriptors below intentionally collapse
	// those findings into abstract review records.
	installSQLPaths := s.discoverPluginSQLPaths(manifest.RootDir, false)
	uninstallSQLPaths := s.discoverPluginSQLPaths(manifest.RootDir, true)
	frontendPagePaths := s.discoverPluginPagePaths(manifest.RootDir)
	frontendSlotPaths := s.discoverPluginSlotPaths(manifest.RootDir)

	descriptors := []*pluginResourceRefDescriptor{
		{
			Kind:      pluginResourceKindManifest,
			Key:       "manifest",
			OwnerType: pluginResourceOwnerTypeFile,
			OwnerKey:  "plugin-manifest",
			Remark:    "One plugin manifest is declared and validated by the host.",
		},
	}

	if normalizePluginType(manifest.Type) == pluginTypeSource {
		descriptors = append(descriptors, &pluginResourceRefDescriptor{
			Kind:      pluginResourceKindBackendEntry,
			Key:       "backend-entry",
			OwnerType: pluginResourceOwnerTypeBackendRegistration,
			OwnerKey:  "source-plugin-backend-entry",
			Remark:    "One source-plugin backend registration entry is compiled into the host binary.",
		})
	}

	if len(installSQLPaths) > 0 {
		descriptors = append(descriptors, &pluginResourceRefDescriptor{
			Kind:      pluginResourceKindInstallSQL,
			Key:       "install-sql-bundle",
			OwnerType: pluginResourceOwnerTypeInstallSQL,
			OwnerKey:  "install-sql-summary",
			Remark:    s.buildPluginResourceSummaryRemark("install SQL assets", len(installSQLPaths)),
		})
	}
	if len(uninstallSQLPaths) > 0 {
		descriptors = append(descriptors, &pluginResourceRefDescriptor{
			Kind:      pluginResourceKindUninstallSQL,
			Key:       "uninstall-sql-bundle",
			OwnerType: pluginResourceOwnerTypeUninstallSQL,
			OwnerKey:  "uninstall-sql-summary",
			Remark:    s.buildPluginResourceSummaryRemark("uninstall SQL assets", len(uninstallSQLPaths)),
		})
	}
	if len(frontendPagePaths) > 0 {
		descriptors = append(descriptors, &pluginResourceRefDescriptor{
			Kind:      pluginResourceKindFrontendPage,
			Key:       "frontend-pages",
			OwnerType: pluginResourceOwnerTypeFrontendPageEntry,
			OwnerKey:  "frontend-page-summary",
			Remark:    s.buildPluginResourceSummaryRemark("frontend page assets", len(frontendPagePaths)),
		})
	}
	if len(frontendSlotPaths) > 0 {
		descriptors = append(descriptors, &pluginResourceRefDescriptor{
			Kind:      pluginResourceKindFrontendSlot,
			Key:       "frontend-slots",
			OwnerType: pluginResourceOwnerTypeFrontendSlotEntry,
			OwnerKey:  "frontend-slot-summary",
			Remark:    s.buildPluginResourceSummaryRemark("frontend slot assets", len(frontendSlotPaths)),
		})
	}

	return descriptors
}

func (s *Service) buildPluginResourceSummaryRemark(resourceLabel string, count int) string {
	return fmt.Sprintf("The host discovered %d %s for the current plugin release.", count, resourceLabel)
}

func (s *Service) buildPluginResourceIdentity(kind string, key string) string {
	return kind + ":" + key
}
