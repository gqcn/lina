// This file materializes dynamic route permissions into hidden synthetic menus
// so the existing sys_menu.perms permission model stays reusable.

package plugin

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

const pluginPermissionMenuNamePrefix = "动态路由权限:"

func (s *Service) buildDynamicRoutePermissionMenuSpecs(manifest *pluginManifest) []*pluginMenuSpec {
	if manifest == nil || len(manifest.Routes) == 0 {
		return []*pluginMenuSpec{}
	}

	items := make([]*pluginMenuSpec, 0)
	seen := make(map[string]struct{})
	for _, route := range manifest.Routes {
		if route == nil || strings.TrimSpace(route.Permission) == "" {
			continue
		}
		permission := strings.TrimSpace(route.Permission)
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		items = append(items, &pluginMenuSpec{
			Key:     buildDynamicRoutePermissionMenuKey(manifest.ID, permission),
			Name:    pluginPermissionMenuNamePrefix + permission,
			Perms:   permission,
			Type:    pluginMenuTypeButton.String(),
			Visible: intPtr(0),
			Status:  intPtr(pluginMenuDefaultStatus),
			Remark:  "plugin:" + manifest.ID + ":dynamic-route-permission",
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Perms < items[j].Perms
	})
	return items
}

func buildDynamicRoutePermissionMenuKey(pluginID string, permission string) string {
	resourceKey := strings.NewReplacer(":", "-", "/", "-", ".", "-").Replace(strings.TrimSpace(permission))
	return pluginMenuKeyPrefix + strings.TrimSpace(pluginID) + ":perm:" + resourceKey
}

func intPtr(value int) *int {
	return &value
}

func (s *Service) syncDynamicRoutePermissionMenus(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}
	permissionMenus := s.buildDynamicRoutePermissionMenuSpecs(manifest)
	if len(permissionMenus) == 0 {
		return nil
	}
	resolvedIDs := make(map[string]int, len(permissionMenus))
	existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
	if err != nil {
		return err
	}
	existingByKey := make(map[string]*entity.SysMenu, len(existingMenus))
	for _, menu := range existingMenus {
		if menu == nil {
			continue
		}
		existingByKey[menu.MenuKey] = menu
	}
	for _, spec := range permissionMenus {
		menuID, err := s.upsertPluginMenu(ctx, spec, 0, existingByKey[spec.Key])
		if err != nil {
			return err
		}
		resolvedIDs[spec.Key] = menuID
	}
	return s.ensurePluginMenuAdminBindings(ctx, resolvedIDs)
}

func (s *Service) syncPluginMenusAndPermissions(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}
	return dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_ = tx
		if err := s.syncPluginMenus(ctx, manifest); err != nil {
			return err
		}
		return s.syncDynamicRoutePermissionMenus(ctx, manifest)
	})
}

func (s *Service) validateDynamicRoutePermissionMenus(manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}
	for _, route := range manifest.Routes {
		if route == nil || strings.TrimSpace(route.Permission) == "" {
			continue
		}
		permission := strings.TrimSpace(route.Permission)
		if !strings.HasPrefix(permission, manifest.ID+":") {
			return gerror.Newf("动态路由 permission 必须使用当前插件前缀: %s", permission)
		}
	}
	return nil
}
