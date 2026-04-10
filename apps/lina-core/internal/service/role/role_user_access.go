package role

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

// UserAccessContext describes the role, menu, and permission data required by the current user session.
type UserAccessContext struct {
	RoleIds      []int    // RoleIds contains all role IDs bound to the user.
	RoleNames    []string // RoleNames contains enabled role names bound to the user.
	MenuIds      []int    // MenuIds contains all menu IDs reachable through the user's roles.
	Permissions  []string // Permissions contains effective button permissions after plugin filtering.
	IsSuperAdmin bool     // IsSuperAdmin reports whether the user owns the built-in admin role.
}

// GetUserAccessContext loads the user's roles, menus, and permissions with one shared role-ID query.
func (s *Service) GetUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error) {
	roleIds, err := s.GetUserRoleIds(ctx, userId)
	if err != nil {
		return nil, err
	}

	roles, err := s.getUserRolesByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}

	menuIds, err := s.getUserMenuIdsByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}

	permissions, err := s.getUserPermissionsByMenuIds(ctx, menuIds)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		roleNames = append(roleNames, role.Name)
	}

	if roleNames == nil {
		roleNames = []string{}
	}
	if menuIds == nil {
		menuIds = []int{}
	}
	if permissions == nil {
		permissions = []string{}
	}

	return &UserAccessContext{
		RoleIds:      roleIds,
		RoleNames:    roleNames,
		MenuIds:      menuIds,
		Permissions:  permissions,
		IsSuperAdmin: hasRoleId(roleIds, 1),
	}, nil
}

func (s *Service) getUserRolesByRoleIds(ctx context.Context, roleIds []int) ([]*entity.SysRole, error) {
	if len(roleIds) == 0 {
		return []*entity.SysRole{}, nil
	}

	var (
		cols  = dao.SysRole.Columns()
		roles []*entity.SysRole
	)

	err := dao.SysRole.Ctx(ctx).
		WhereIn(cols.Id, roleIds).
		Where(cols.Status, 1).
		Scan(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *Service) getUserMenuIdsByRoleIds(ctx context.Context, roleIds []int) ([]int, error) {
	if len(roleIds) == 0 {
		return []int{}, nil
	}

	var (
		rmCols    = dao.SysRoleMenu.Columns()
		roleMenus []*entity.SysRoleMenu
	)

	err := dao.SysRoleMenu.Ctx(ctx).
		WhereIn(rmCols.RoleId, roleIds).
		Scan(&roleMenus)
	if err != nil {
		return nil, err
	}

	menuIds := make([]int, 0, len(roleMenus))
	menuIdSet := make(map[int]bool, len(roleMenus))
	for _, roleMenu := range roleMenus {
		if roleMenu == nil || menuIdSet[roleMenu.MenuId] {
			continue
		}
		menuIds = append(menuIds, roleMenu.MenuId)
		menuIdSet[roleMenu.MenuId] = true
	}
	return menuIds, nil
}

func (s *Service) getUserPermissionsByMenuIds(ctx context.Context, menuIds []int) ([]string, error) {
	if len(menuIds) == 0 {
		return []string{}, nil
	}

	var (
		menuCols = dao.SysMenu.Columns()
		menus    []*entity.SysMenu
	)

	err := dao.SysMenu.Ctx(ctx).
		WhereIn(menuCols.Id, menuIds).
		Where(menuCols.Type, "B").
		Where(menuCols.Status, 1).
		Scan(&menus)
	if err != nil {
		return nil, err
	}

	menus = s.pluginSvc.FilterPermissionMenus(ctx, menus)

	perms := make([]string, 0, len(menus))
	for _, menu := range menus {
		if menu == nil || menu.Perms == "" {
			continue
		}
		perms = append(perms, menu.Perms)
	}
	return perms, nil
}

func hasRoleId(roleIds []int, roleId int) bool {
	for _, currentRoleID := range roleIds {
		if currentRoleID == roleId {
			return true
		}
	}
	return false
}
