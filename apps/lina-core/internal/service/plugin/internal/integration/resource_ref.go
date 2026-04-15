// This file synchronizes abstract plugin resource descriptors into
// sys_plugin_resource_ref for framework-agnostic governance review.

package integration

import (
	"context"
	"fmt"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
)

// SyncPluginResourceReferences keeps sys_plugin_resource_ref aligned with the current
// abstract resource summary for the given manifest.
// It implements catalog.ResourceRefSyncer.
func (s *Service) SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error {
	if manifest == nil {
		return nil
	}

	release, err := s.catalogSvc.GetRelease(ctx, manifest.ID, manifest.Version)
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
		existingMap[buildPluginResourceIdentity(item.ResourceType, item.ResourceKey)] = item
	}

	seen := make(map[string]struct{})
	for _, descriptor := range s.buildPluginResourceRefDescriptors(manifest) {
		identity := buildPluginResourceIdentity(descriptor.Kind.String(), descriptor.Key)
		seen[identity] = struct{}{}

		if existing, ok := existingMap[identity]; ok {
			// Only update abstract ownership and review remarks. Concrete file paths are
			// deliberately excluded so the schema stays framework-agnostic. Runtime
			// uninstall currently soft-deletes old rows, so repeated sync must also be
			// able to revive matching identities instead of colliding with the table
			// unique key on a fresh insert.
			data := do.SysPluginResourceRef{
				OwnerType: descriptor.OwnerType.String(),
				OwnerKey:  descriptor.OwnerKey,
				Remark:    descriptor.Remark,
			}
			_, err = dao.SysPluginResourceRef.Ctx(ctx).
				Unscoped().
				Where(do.SysPluginResourceRef{Id: existing.Id}).
				Data(data).
				Update()
			if err != nil {
				return err
			}
			if existing.DeletedAt != nil {
				if _, err = dao.SysPluginResourceRef.Ctx(ctx).
					Unscoped().
					Where(do.SysPluginResourceRef{Id: existing.Id}).
					Data("deleted_at", nil).
					Update(); err != nil {
					return err
				}
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
		identity := buildPluginResourceIdentity(item.ResourceType, item.ResourceKey)
		if _, ok := seen[identity]; ok {
			continue
		}
		if _, err = dao.SysPluginResourceRef.Ctx(ctx).
			Unscoped().
			Where(do.SysPluginResourceRef{Id: item.Id}).
			Delete(); err != nil {
			return err
		}
	}

	return nil
}

// listPluginResourceRefs returns all resource-ref rows for one plugin release, including soft-deleted rows.
func (s *Service) listPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error) {
	items := make([]*entity.SysPluginResourceRef, 0)
	err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{
			PluginId:  pluginID,
			ReleaseId: releaseID,
		}).
		Scan(&items)
	return items, err
}

// buildPluginResourceRefDescriptors converts concrete discovery results into framework-agnostic review records.
func (s *Service) buildPluginResourceRefDescriptors(manifest *catalog.Manifest) []*catalog.ResourceRefDescriptor {
	if manifest == nil {
		return []*catalog.ResourceRefDescriptor{}
	}

	installSQLCount := s.countPluginInstallSQLAssets(manifest)
	uninstallSQLCount := s.countPluginUninstallSQLAssets(manifest)
	frontendPagePaths := s.catalogSvc.ListFrontendPagePaths(manifest)
	frontendSlotPaths := s.catalogSvc.ListFrontendSlotPaths(manifest)

	descriptors := []*catalog.ResourceRefDescriptor{
		{
			Kind:      catalog.ResourceKindManifest,
			Key:       "manifest",
			OwnerType: catalog.ResourceOwnerTypeFile,
			OwnerKey:  "plugin-manifest",
			Remark:    "One plugin manifest is declared and validated by the host.",
		},
	}

	if catalog.NormalizeType(manifest.Type) == catalog.TypeSource {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindBackendEntry,
			Key:       "backend-entry",
			OwnerType: catalog.ResourceOwnerTypeBackendRegistration,
			OwnerKey:  "source-plugin-backend-entry",
			Remark:    "One source-plugin backend registration entry is compiled into the host binary.",
		})
	} else if manifest.RuntimeArtifact != nil {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindRuntimeWasm,
			Key:       "runtime-wasm-artifact",
			OwnerType: catalog.ResourceOwnerTypeRuntimeArtifact,
			OwnerKey:  "runtime-wasm-artifact",
			Remark:    buildRuntimeArtifactRemark(manifest),
		})
		if manifest.RuntimeArtifact.FrontendAssetCount > 0 {
			descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
				Kind:      catalog.ResourceKindRuntimeFrontend,
				Key:       "runtime-frontend-assets",
				OwnerType: catalog.ResourceOwnerTypeRuntimeFrontend,
				OwnerKey:  "runtime-frontend-assets",
				Remark:    buildPluginResourceSummaryRemark("runtime frontend assets", manifest.RuntimeArtifact.FrontendAssetCount),
			})
		}
	}

	if installSQLCount > 0 {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindInstallSQL,
			Key:       "install-sql-bundle",
			OwnerType: catalog.ResourceOwnerTypeInstallSQL,
			OwnerKey:  "install-sql-summary",
			Remark:    buildPluginResourceSummaryRemark("install SQL assets", installSQLCount),
		})
	}
	if uninstallSQLCount > 0 {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindUninstallSQL,
			Key:       "uninstall-sql-bundle",
			OwnerType: catalog.ResourceOwnerTypeUninstallSQL,
			OwnerKey:  "uninstall-sql-summary",
			Remark:    buildPluginResourceSummaryRemark("uninstall SQL assets", uninstallSQLCount),
		})
	}
	if len(frontendPagePaths) > 0 {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindFrontendPage,
			Key:       "frontend-pages",
			OwnerType: catalog.ResourceOwnerTypeFrontendPageEntry,
			OwnerKey:  "frontend-page-summary",
			Remark:    buildPluginResourceSummaryRemark("frontend page assets", len(frontendPagePaths)),
		})
	}
	if len(frontendSlotPaths) > 0 {
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindFrontendSlot,
			Key:       "frontend-slots",
			OwnerType: catalog.ResourceOwnerTypeFrontendSlotEntry,
			OwnerKey:  "frontend-slot-summary",
			Remark:    buildPluginResourceSummaryRemark("frontend slot assets", len(frontendSlotPaths)),
		})
	}
	for _, menu := range manifest.Menus {
		if menu == nil || strings.TrimSpace(menu.Key) == "" {
			continue
		}
		descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
			Kind:      catalog.ResourceKindMenu,
			Key:       strings.TrimSpace(menu.Key),
			OwnerType: catalog.ResourceOwnerTypeMenuEntry,
			OwnerKey:  "manifest-menu",
			Remark:    buildPluginMenuResourceRemark(menu),
		})
	}

	descriptors = appendHostServiceResourceDescriptors(descriptors, manifest.HostServices)

	return descriptors
}

// countPluginInstallSQLAssets returns the number of install SQL steps for the manifest.
// For dynamic plugins the count comes from the embedded artifact; for source plugins it scans disk.
func (s *Service) countPluginInstallSQLAssets(manifest *catalog.Manifest) int {
	if manifest == nil {
		return 0
	}
	if manifest.RuntimeArtifact != nil {
		return len(manifest.RuntimeArtifact.InstallSQLAssets)
	}
	return len(s.catalogSvc.ListInstallSQLPaths(manifest))
}

// countPluginUninstallSQLAssets returns the number of uninstall SQL steps for the manifest.
// For dynamic plugins the count comes from the embedded artifact; for source plugins it scans disk.
func (s *Service) countPluginUninstallSQLAssets(manifest *catalog.Manifest) int {
	if manifest == nil {
		return 0
	}
	if manifest.RuntimeArtifact != nil {
		return len(manifest.RuntimeArtifact.UninstallSQLAssets)
	}
	return len(s.catalogSvc.ListUninstallSQLPaths(manifest))
}

// buildPluginResourceSummaryRemark formats the standard governance discovery summary line.
func buildPluginResourceSummaryRemark(resourceLabel string, count int) string {
	return fmt.Sprintf("The host discovered %d %s for the current plugin release.", count, resourceLabel)
}

// buildPluginResourceIdentity returns a stable composite key for one resource ref row.
func buildPluginResourceIdentity(kind string, key string) string {
	return kind + ":" + key
}

// buildRuntimeArtifactRemark summarizes runtime WASM metadata for governance review.
// Inlined from runtime/artifact.go to avoid a circular import (integration cannot import runtime).
func buildRuntimeArtifactRemark(manifest *catalog.Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return fmt.Sprintf(
		"The host validated one %s runtime artifact using ABI %s with %d embedded frontend assets, %d install SQL assets, %d uninstall SQL assets, and %d dynamic routes declared.",
		manifest.RuntimeArtifact.RuntimeKind,
		manifest.RuntimeArtifact.ABIVersion,
		manifest.RuntimeArtifact.FrontendAssetCount,
		len(manifest.RuntimeArtifact.InstallSQLAssets),
		len(manifest.RuntimeArtifact.UninstallSQLAssets),
		len(manifest.RuntimeArtifact.RouteContracts),
	)
}

// buildPluginMenuResourceRemark formats the governance remark for one manifest-declared menu entry.
func buildPluginMenuResourceRemark(menu *catalog.MenuSpec) string {
	if menu == nil {
		return "The host discovered one manifest-declared plugin menu."
	}
	return fmt.Sprintf(
		"The host discovered one manifest-declared plugin menu named %q with type %s.",
		strings.TrimSpace(menu.Name),
		normalizeMenuType(menu.Type),
	)
}

func appendHostServiceResourceDescriptors(
	descriptors []*catalog.ResourceRefDescriptor,
	hostServices []*pluginbridge.HostServiceSpec,
) []*catalog.ResourceRefDescriptor {
	if len(hostServices) == 0 {
		return descriptors
	}

	seen := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor == nil {
			continue
		}
		seen[buildPluginResourceIdentity(descriptor.Kind.String(), descriptor.Key)] = struct{}{}
	}

	for _, service := range hostServices {
		if service == nil {
			continue
		}
		kind := mapHostServiceResourceKind(service.Service)
		if kind == "" {
			continue
		}
		if len(service.Tables) > 0 {
			for _, table := range service.Tables {
				normalizedTable := strings.TrimSpace(table)
				if normalizedTable == "" {
					continue
				}
				identity := buildPluginResourceIdentity(kind.String(), normalizedTable)
				if _, ok := seen[identity]; ok {
					continue
				}
				seen[identity] = struct{}{}
				descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
					Kind:      kind,
					Key:       normalizedTable,
					OwnerType: catalog.ResourceOwnerTypeHostServiceResource,
					OwnerKey:  service.Service,
					Remark:    buildHostServiceTableRemark(service.Service, normalizedTable, service.Methods),
				})
			}
			continue
		}
		if len(service.Paths) > 0 {
			for _, item := range service.Paths {
				normalizedPath := strings.TrimSpace(item)
				if normalizedPath == "" {
					continue
				}
				identity := buildPluginResourceIdentity(kind.String(), normalizedPath)
				if _, ok := seen[identity]; ok {
					continue
				}
				seen[identity] = struct{}{}
				descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
					Kind:      kind,
					Key:       normalizedPath,
					OwnerType: catalog.ResourceOwnerTypeHostServiceResource,
					OwnerKey:  service.Service,
					Remark:    buildHostServicePathRemark(service.Service, normalizedPath, service.Methods),
				})
			}
			continue
		}
		if len(service.Resources) == 0 {
			continue
		}
		for _, resource := range service.Resources {
			if resource == nil || strings.TrimSpace(resource.Ref) == "" {
				continue
			}
			identity := buildPluginResourceIdentity(kind.String(), strings.TrimSpace(resource.Ref))
			if _, ok := seen[identity]; ok {
				continue
			}
			seen[identity] = struct{}{}
			descriptors = append(descriptors, &catalog.ResourceRefDescriptor{
				Kind:      kind,
				Key:       strings.TrimSpace(resource.Ref),
				OwnerType: catalog.ResourceOwnerTypeHostServiceResource,
				OwnerKey:  service.Service,
				Remark:    buildHostServiceResourceRemark(service.Service, resource.Ref, service.Methods),
			})
		}
	}

	return descriptors
}

func mapHostServiceResourceKind(service string) catalog.ResourceKind {
	switch strings.TrimSpace(service) {
	case pluginbridge.HostServiceStorage:
		return catalog.ResourceKindHostStorage
	case pluginbridge.HostServiceNetwork:
		return catalog.ResourceKindHostUpstream
	case pluginbridge.HostServiceData:
		return catalog.ResourceKindHostData
	case pluginbridge.HostServiceCache:
		return catalog.ResourceKindHostCache
	case pluginbridge.HostServiceLock:
		return catalog.ResourceKindHostLock
	case pluginbridge.HostServiceSecret:
		return catalog.ResourceKindHostSecret
	case pluginbridge.HostServiceEvent:
		return catalog.ResourceKindHostEventTopic
	case pluginbridge.HostServiceQueue:
		return catalog.ResourceKindHostQueue
	case pluginbridge.HostServiceNotify:
		return catalog.ResourceKindHostNotify
	default:
		return ""
	}
}

func buildHostServiceResourceRemark(service string, ref string, methods []string) string {
	methodSummary := "no methods"
	if len(methods) > 0 {
		methodSummary = strings.Join(methods, ", ")
	}
	return fmt.Sprintf(
		"The host discovered one governed host service resource ref %q for service %s with methods [%s].",
		strings.TrimSpace(ref),
		strings.TrimSpace(service),
		methodSummary,
	)
}

func buildHostServicePathRemark(service string, storagePath string, methods []string) string {
	methodSummary := "no methods"
	if len(methods) > 0 {
		methodSummary = strings.Join(methods, ", ")
	}
	return fmt.Sprintf(
		"The host discovered one governed host service path %q for service %s with methods [%s].",
		strings.TrimSpace(storagePath),
		strings.TrimSpace(service),
		methodSummary,
	)
}

func buildHostServiceTableRemark(service string, table string, methods []string) string {
	methodSummary := "no methods"
	if len(methods) > 0 {
		methodSummary = strings.Join(methods, ", ")
	}
	return fmt.Sprintf(
		"The host discovered one governed host service table %q for service %s with methods [%s].",
		strings.TrimSpace(table),
		strings.TrimSpace(service),
		methodSummary,
	)
}

// ListPluginResourceRefs is the exported form of listPluginResourceRefs for cross-package access.
func (s *Service) ListPluginResourceRefs(ctx context.Context, pluginID string, releaseID int) ([]*entity.SysPluginResourceRef, error) {
	return s.listPluginResourceRefs(ctx, pluginID, releaseID)
}

// BuildResourceRefDescriptors is the exported form of buildPluginResourceRefDescriptors for cross-package access.
func (s *Service) BuildResourceRefDescriptors(manifest *catalog.Manifest) []*catalog.ResourceRefDescriptor {
	return s.buildPluginResourceRefDescriptors(manifest)
}
