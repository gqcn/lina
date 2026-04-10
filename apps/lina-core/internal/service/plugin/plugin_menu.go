// This file defines manifest-driven plugin menu metadata together with the
// host-side synchronization logic that writes and removes plugin menus.

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

type pluginMenuType string

const (
	pluginMenuTypeDirectory pluginMenuType = "D"
	pluginMenuTypePage      pluginMenuType = "M"
	pluginMenuTypeButton    pluginMenuType = "B"

	pluginMenuDefaultVisible = 1
	pluginMenuDefaultStatus  = 1
	pluginMenuDefaultIsFrame = 0
	pluginMenuDefaultIsCache = 0
	pluginDefaultAdminRoleID = 1
)

// pluginMenuSpec defines one manifest-driven host menu declaration.
type pluginMenuSpec struct {
	Key        string                 `yaml:"key" json:"key"`
	ParentKey  string                 `yaml:"parent_key,omitempty" json:"parent_key,omitempty"`
	Name       string                 `yaml:"name" json:"name"`
	Path       string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Component  string                 `yaml:"component,omitempty" json:"component,omitempty"`
	Perms      string                 `yaml:"perms,omitempty" json:"perms,omitempty"`
	Icon       string                 `yaml:"icon,omitempty" json:"icon,omitempty"`
	Type       string                 `yaml:"type,omitempty" json:"type,omitempty"`
	Sort       int                    `yaml:"sort,omitempty" json:"sort,omitempty"`
	Visible    *int                   `yaml:"visible,omitempty" json:"visible,omitempty"`
	Status     *int                   `yaml:"status,omitempty" json:"status,omitempty"`
	IsFrame    *int                   `yaml:"is_frame,omitempty" json:"is_frame,omitempty"`
	IsCache    *int                   `yaml:"is_cache,omitempty" json:"is_cache,omitempty"`
	Query      map[string]interface{} `yaml:"query,omitempty" json:"query,omitempty"`
	QueryParam string                 `yaml:"query_param,omitempty" json:"query_param,omitempty"`
	Remark     string                 `yaml:"remark,omitempty" json:"remark,omitempty"`
}

// syncPluginMenus reconciles one plugin's declared menus into sys_menu and
// guarantees that the default administrator role keeps access to them.
func (s *Service) syncPluginMenus(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}

	return dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_ = tx
		declaredKeys := s.listDeclaredPluginMenuKeys(manifest)
		existingMenus, err := s.listPluginMenusByPlugin(ctx, manifest.ID)
		if err != nil {
			return err
		}

		existingByKey := make(map[string]*entity.SysMenu, len(existingMenus))
		staleKeys := make([]string, 0)
		for _, item := range existingMenus {
			if item == nil {
				continue
			}
			existingByKey[item.MenuKey] = item
			if _, ok := declaredKeys[item.MenuKey]; !ok {
				staleKeys = append(staleKeys, item.MenuKey)
			}
		}

		externalParents, err := s.listPluginMenuExternalParents(ctx, manifest)
		if err != nil {
			return err
		}

		resolvedIDs := make(map[string]int, len(manifest.Menus))
		pendingMenus := append([]*pluginMenuSpec(nil), manifest.Menus...)
		for len(pendingMenus) > 0 {
			nextPending := make([]*pluginMenuSpec, 0, len(pendingMenus))
			progressed := false

			for _, spec := range pendingMenus {
				if spec == nil {
					continue
				}

				parentID, resolved, err := s.resolvePluginMenuParentID(spec, declaredKeys, resolvedIDs, externalParents)
				if err != nil {
					return err
				}
				if !resolved {
					nextPending = append(nextPending, spec)
					continue
				}

				menuID, err := s.upsertPluginMenu(ctx, spec, parentID, existingByKey[spec.Key])
				if err != nil {
					return err
				}
				resolvedIDs[spec.Key] = menuID
				progressed = true
			}

			if !progressed {
				unresolved := make([]string, 0, len(nextPending))
				for _, spec := range nextPending {
					if spec == nil {
						continue
					}
					unresolved = append(unresolved, spec.Key)
				}
				sort.Strings(unresolved)
				return gerror.Newf("插件菜单 parent_key 无法解析: %s", strings.Join(unresolved, ", "))
			}

			pendingMenus = nextPending
		}

		if err := s.ensurePluginMenuAdminBindings(ctx, resolvedIDs); err != nil {
			return err
		}
		return s.deletePluginMenusByKeys(ctx, staleKeys)
	})
}

// deletePluginMenusByManifest removes plugin menus declared by manifest and
// clears their role bindings during uninstall.
func (s *Service) deletePluginMenusByManifest(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil {
		return nil
	}

	menuKeys := make([]string, 0, len(manifest.Menus))
	for _, spec := range manifest.Menus {
		if spec == nil || strings.TrimSpace(spec.Key) == "" {
			continue
		}
		menuKeys = append(menuKeys, strings.TrimSpace(spec.Key))
	}
	return s.deletePluginMenusByKeys(ctx, menuKeys)
}

func (s *Service) validatePluginManifestMenus(manifest *pluginManifest) error {
	if manifest == nil || len(manifest.Menus) == 0 {
		return nil
	}

	declaredKeys := make(map[string]struct{}, len(manifest.Menus))
	for index, spec := range manifest.Menus {
		if spec == nil {
			return gerror.Newf("第 %d 个菜单声明不能为空", index+1)
		}

		spec.Key = strings.TrimSpace(spec.Key)
		spec.ParentKey = strings.TrimSpace(spec.ParentKey)
		spec.Name = strings.TrimSpace(spec.Name)
		spec.Path = strings.TrimSpace(spec.Path)
		spec.Component = strings.TrimSpace(spec.Component)
		spec.Perms = strings.TrimSpace(spec.Perms)
		spec.Icon = strings.TrimSpace(spec.Icon)
		spec.Type = normalizePluginMenuType(spec.Type).String()
		spec.QueryParam = strings.TrimSpace(spec.QueryParam)
		spec.Remark = strings.TrimSpace(spec.Remark)

		if spec.Key == "" {
			return gerror.Newf("第 %d 个菜单声明缺少 key", index+1)
		}
		if spec.Name == "" {
			return gerror.Newf("插件菜单缺少 name: %s", spec.Key)
		}
		if !isSupportedPluginMenuType(normalizePluginMenuType(spec.Type)) {
			return gerror.Newf("插件菜单类型仅支持 D/M/B: %s", spec.Key)
		}
		if spec.ParentKey == spec.Key {
			return gerror.Newf("插件菜单 parent_key 不能指向自己: %s", spec.Key)
		}
		pluginID := s.parsePluginIDFromTaggedValue(spec.Key, pluginMenuKeyPrefix)
		if pluginID == "" || pluginID != manifest.ID {
			return gerror.Newf("插件菜单 key 必须使用当前插件前缀 plugin:%s:* : %s", manifest.ID, spec.Key)
		}
		if parentPluginID := s.parsePluginIDFromTaggedValue(spec.ParentKey, pluginMenuKeyPrefix); parentPluginID != "" && parentPluginID != manifest.ID {
			return gerror.Newf("插件菜单 parent_key 不允许引用其他插件菜单: %s -> %s", spec.Key, spec.ParentKey)
		}
		if _, ok := declaredKeys[spec.Key]; ok {
			return gerror.Newf("插件菜单 key 重复: %s", spec.Key)
		}
		declaredKeys[spec.Key] = struct{}{}

		if _, err := normalizePluginMenuFlag(spec.Visible, pluginMenuDefaultVisible); err != nil {
			return gerror.Wrapf(err, "插件菜单 visible 非法: %s", spec.Key)
		}
		if _, err := normalizePluginMenuFlag(spec.Status, pluginMenuDefaultStatus); err != nil {
			return gerror.Wrapf(err, "插件菜单 status 非法: %s", spec.Key)
		}
		if _, err := normalizePluginMenuFlag(spec.IsFrame, pluginMenuDefaultIsFrame); err != nil {
			return gerror.Wrapf(err, "插件菜单 is_frame 非法: %s", spec.Key)
		}
		if _, err := normalizePluginMenuFlag(spec.IsCache, pluginMenuDefaultIsCache); err != nil {
			return gerror.Wrapf(err, "插件菜单 is_cache 非法: %s", spec.Key)
		}
		if _, err := buildPluginMenuQueryParam(spec); err != nil {
			return gerror.Wrapf(err, "插件菜单 query 非法: %s", spec.Key)
		}
	}

	for _, spec := range manifest.Menus {
		if spec == nil || spec.ParentKey == "" {
			continue
		}
		parentPluginID := s.parsePluginIDFromTaggedValue(spec.ParentKey, pluginMenuKeyPrefix)
		if parentPluginID != manifest.ID {
			continue
		}
		if _, ok := declaredKeys[spec.ParentKey]; !ok {
			return gerror.Newf("插件菜单引用了未声明的 parent_key: %s -> %s", spec.Key, spec.ParentKey)
		}
	}

	return nil
}

func normalizePluginMenuType(value string) pluginMenuType {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "":
		return pluginMenuTypePage
	case pluginMenuTypeDirectory.String():
		return pluginMenuTypeDirectory
	case pluginMenuTypePage.String():
		return pluginMenuTypePage
	case pluginMenuTypeButton.String():
		return pluginMenuTypeButton
	default:
		return ""
	}
}

func isSupportedPluginMenuType(value pluginMenuType) bool {
	return value == pluginMenuTypeDirectory || value == pluginMenuTypePage || value == pluginMenuTypeButton
}

func normalizePluginMenuFlag(value *int, defaultValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	if *value != 0 && *value != 1 {
		return 0, gerror.New("仅支持 0 或 1")
	}
	return *value, nil
}

func buildPluginMenuQueryParam(spec *pluginMenuSpec) (string, error) {
	if spec == nil {
		return "", nil
	}
	if strings.TrimSpace(spec.QueryParam) != "" {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(spec.QueryParam), &payload); err != nil {
			return "", err
		}
		if len(payload) == 0 {
			return "", nil
		}
		content, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	if len(spec.Query) == 0 {
		return "", nil
	}
	content, err := json.Marshal(spec.Query)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *Service) listDeclaredPluginMenuKeys(manifest *pluginManifest) map[string]struct{} {
	declaredKeys := make(map[string]struct{}, len(manifest.Menus))
	if manifest == nil {
		return declaredKeys
	}
	for _, spec := range manifest.Menus {
		if spec == nil || strings.TrimSpace(spec.Key) == "" {
			continue
		}
		declaredKeys[strings.TrimSpace(spec.Key)] = struct{}{}
	}
	return declaredKeys
}

func (s *Service) listPluginMenuExternalParents(ctx context.Context, manifest *pluginManifest) (map[string]*entity.SysMenu, error) {
	declaredKeys := s.listDeclaredPluginMenuKeys(manifest)
	parentKeys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, spec := range manifest.Menus {
		if spec == nil || spec.ParentKey == "" {
			continue
		}
		if _, ok := declaredKeys[spec.ParentKey]; ok {
			continue
		}
		if _, ok := seen[spec.ParentKey]; ok {
			continue
		}
		seen[spec.ParentKey] = struct{}{}
		parentKeys = append(parentKeys, spec.ParentKey)
	}
	return s.listMenusByKeys(ctx, parentKeys, false)
}

func (s *Service) resolvePluginMenuParentID(
	spec *pluginMenuSpec,
	declaredKeys map[string]struct{},
	resolvedIDs map[string]int,
	externalParents map[string]*entity.SysMenu,
) (int, bool, error) {
	if spec == nil || strings.TrimSpace(spec.ParentKey) == "" {
		return 0, true, nil
	}

	parentKey := strings.TrimSpace(spec.ParentKey)
	if _, ok := declaredKeys[parentKey]; ok {
		parentID, resolved := resolvedIDs[parentKey]
		return parentID, resolved, nil
	}

	parent, ok := externalParents[parentKey]
	if !ok || parent == nil {
		return 0, false, gerror.Newf("插件菜单 parent_key 不存在: %s -> %s", spec.Key, spec.ParentKey)
	}
	return parent.Id, true, nil
}

func (s *Service) upsertPluginMenu(
	ctx context.Context,
	spec *pluginMenuSpec,
	parentID int,
	existing *entity.SysMenu,
) (int, error) {
	if spec == nil {
		return 0, gerror.New("插件菜单声明不能为空")
	}

	queryParam, err := buildPluginMenuQueryParam(spec)
	if err != nil {
		return 0, err
	}
	visible, err := normalizePluginMenuFlag(spec.Visible, pluginMenuDefaultVisible)
	if err != nil {
		return 0, err
	}
	status, err := normalizePluginMenuFlag(spec.Status, pluginMenuDefaultStatus)
	if err != nil {
		return 0, err
	}
	isFrame, err := normalizePluginMenuFlag(spec.IsFrame, pluginMenuDefaultIsFrame)
	if err != nil {
		return 0, err
	}
	isCache, err := normalizePluginMenuFlag(spec.IsCache, pluginMenuDefaultIsCache)
	if err != nil {
		return 0, err
	}

	if existing != nil && existing.DeletedAt != nil {
		if _, err = dao.SysMenu.Ctx(ctx).
			Unscoped().
			Where(do.SysMenu{Id: existing.Id}).
			Delete(); err != nil {
			return 0, err
		}
		existing = nil
	}

	data := do.SysMenu{
		ParentId:   parentID,
		MenuKey:    spec.Key,
		Name:       spec.Name,
		Path:       spec.Path,
		Component:  spec.Component,
		Perms:      spec.Perms,
		Icon:       spec.Icon,
		Type:       normalizePluginMenuType(spec.Type).String(),
		Sort:       spec.Sort,
		Visible:    visible,
		Status:     status,
		IsFrame:    isFrame,
		IsCache:    isCache,
		QueryParam: queryParam,
		Remark:     spec.Remark,
	}

	if existing == nil {
		menuID, err := dao.SysMenu.Ctx(ctx).Data(data).InsertAndGetId()
		if err != nil {
			return 0, err
		}
		return int(menuID), nil
	}

	if _, err = dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{Id: existing.Id}).
		Data(data).
		Update(); err != nil {
		return 0, err
	}
	return existing.Id, nil
}

func (s *Service) ensurePluginMenuAdminBindings(ctx context.Context, resolvedIDs map[string]int) error {
	menuIDs := make([]int, 0, len(resolvedIDs))
	for _, menuID := range resolvedIDs {
		if menuID <= 0 {
			continue
		}
		menuIDs = append(menuIDs, menuID)
	}
	sort.Ints(menuIDs)

	for _, menuID := range menuIDs {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			Data(do.SysRoleMenu{
				RoleId: pluginDefaultAdminRoleID,
				MenuId: menuID,
			}).
			Save(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) listPluginMenusByPlugin(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	pattern := fmt.Sprintf("%s%s:%%", pluginMenuKeyPrefix, strings.TrimSpace(pluginID))
	cols := dao.SysMenu.Columns()
	items := make([]*entity.SysMenu, 0)
	err := dao.SysMenu.Ctx(ctx).
		Unscoped().
		WhereLike(cols.MenuKey, pattern).
		Order(cols.Id + " ASC").
		Scan(&items)
	return items, err
}

func (s *Service) listMenusByKeys(ctx context.Context, menuKeys []string, unscoped bool) (map[string]*entity.SysMenu, error) {
	result := make(map[string]*entity.SysMenu, len(menuKeys))
	if len(menuKeys) == 0 {
		return result, nil
	}

	m := dao.SysMenu.Ctx(ctx)
	if unscoped {
		m = m.Unscoped()
	}

	cols := dao.SysMenu.Columns()
	items := make([]*entity.SysMenu, 0)
	if err := m.WhereIn(cols.MenuKey, menuKeys).Order(cols.Id + " ASC").Scan(&items); err != nil {
		return nil, err
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		result[item.MenuKey] = item
	}
	return result, nil
}

func (s *Service) deletePluginMenusByKeys(ctx context.Context, menuKeys []string) error {
	if len(menuKeys) == 0 {
		return nil
	}

	menuMap, err := s.listMenusByKeys(ctx, menuKeys, true)
	if err != nil {
		return err
	}

	menuIDs := make([]int, 0, len(menuMap))
	for _, item := range menuMap {
		if item == nil {
			continue
		}
		menuIDs = append(menuIDs, item.Id)
	}
	sort.Ints(menuIDs)

	if len(menuIDs) > 0 {
		menuIDValues := make([]interface{}, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			menuIDValues = append(menuIDValues, menuID)
		}
		if _, err = dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDValues).
			Delete(); err != nil {
			return err
		}
	}

	if _, err = dao.SysMenu.Ctx(ctx).
		Unscoped().
		WhereIn(dao.SysMenu.Columns().MenuKey, menuKeys).
		Delete(); err != nil {
		return err
	}
	return nil
}

func (value pluginMenuType) String() string {
	return string(value)
}
