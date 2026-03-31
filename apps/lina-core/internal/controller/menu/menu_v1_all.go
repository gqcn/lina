package menu

import (
	"context"

	v1 "lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

// GetAll returns all menus for the current user in Vben route format
func (c *ControllerV1) GetAll(ctx context.Context, req *v1.GetAllReq) (res *v1.GetAllRes, err error) {
	// Get user ID from business context (set by auth middleware)
	bizCtx := c.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return &v1.GetAllRes{List: []*v1.MenuRouteItem{}}, nil
	}
	userId := bizCtx.UserId

	// Check if super admin
	isSuperAdmin := c.roleSvc.IsSuperAdmin(ctx, userId)

	var menuTree []*menusvc.MenuItem

	statusNormal := 1
	if isSuperAdmin {
		// Super admin gets all enabled menus
		allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
			Status: &statusNormal,
		})
		if err != nil {
			return nil, err
		}
		menuTree = c.menuSvc.BuildTree(allMenus.List)
	} else {
		// Regular user gets menus based on roles
		menuIds, err := c.roleSvc.GetUserMenuIds(ctx, userId)
		if err != nil {
			return nil, err
		}
		if len(menuIds) > 0 {
			allMenus, err := c.menuSvc.List(ctx, menusvc.ListInput{
				Status: &statusNormal,
			})
			if err != nil {
				return nil, err
			}
			// Filter menus by user's menu IDs
			menuMap := make(map[int]bool)
			for _, id := range menuIds {
				menuMap[id] = true
			}
			filteredMenus := make([]*menusvc.MenuItem, 0)
			for _, m := range allMenus.List {
				if menuMap[m.Id] {
					filteredMenus = append(filteredMenus, &menusvc.MenuItem{
						Id:         m.Id,
						ParentId:   m.ParentId,
						Name:       m.Name,
						Path:       m.Path,
						Component:  m.Component,
						Perms:      m.Perms,
						Icon:       m.Icon,
						Type:       m.Type,
						Sort:       m.Sort,
						Visible:    m.Visible,
						Status:     m.Status,
						IsFrame:    m.IsFrame,
						IsCache:    m.IsCache,
						QueryParam: m.QueryParam,
						Remark:     m.Remark,
						CreatedAt:  m.CreatedAt.String(),
						UpdatedAt:  m.UpdatedAt.String(),
						Children:   make([]*menusvc.MenuItem, 0),
					})
				}
			}
			menuTree = buildFilteredTree(filteredMenus)
		}
	}

	// Convert to Vben route format
	routes := convertToRouteItems(menuTree)

	return &v1.GetAllRes{List: routes}, nil
}

// getUserIdFromContext extracts user ID from context
func getUserIdFromContext(ctx context.Context) int {
	val := ctx.Value("userId")
	if val == nil {
		return 0
	}
	userId, ok := val.(int)
	if !ok {
		return 0
	}
	return userId
}

// buildFilteredTree builds a tree from filtered menu items
func buildFilteredTree(items []*menusvc.MenuItem) []*menusvc.MenuItem {
	nodeMap := make(map[int]*menusvc.MenuItem)
	for _, m := range items {
		nodeMap[m.Id] = m
	}

	var roots []*menusvc.MenuItem
	for _, m := range items {
		if m.ParentId == 0 {
			roots = append(roots, m)
		} else {
			if parent, ok := nodeMap[m.ParentId]; ok {
				parent.Children = append(parent.Children, m)
			}
		}
	}
	return roots
}

// convertToRouteItems converts menu items to Vben route format
func convertToRouteItems(items []*menusvc.MenuItem) []*v1.MenuRouteItem {
	result := make([]*v1.MenuRouteItem, 0, len(items))
	for _, item := range items {
		route := &v1.MenuRouteItem{
			Id:       item.Id,
			ParentId: item.ParentId,
			Name:     generateRouteName(item),
			Path:     generateRoutePath(item),
			Meta: &v1.MenuRouteMeta{
				Title:            item.Name,
				Icon:             item.Icon,
				HideInMenu:       item.Visible == 0,
				KeepAlive:        item.IsCache == 1,
				Order:            item.Sort,
				Authority:        item.Perms,
				IgnoreAccess:     false,
				HideInBreadcrumb: false,
				HideInTab:        false,
				ActiveIcon:       "",
			},
		}

		// Set component for menu type (M) - actual pages
		if item.Type == "M" {
			route.Component = generateComponentPath(item.Component)
		}

		// Set redirect for directory type (D) with children
		if item.Type == "D" && len(item.Children) > 0 {
			// Redirect to first child
			if item.Children[0].Type == "M" {
				route.Redirect = generateRoutePath(item.Children[0])
			}
		}

		// Convert children recursively
		if len(item.Children) > 0 {
			route.Children = convertToRouteItems(item.Children)
		}

		result = append(result, route)
	}
	return result
}

// generateRouteName generates route name from menu
func generateRouteName(item *menusvc.MenuItem) string {
	if item.Path != "" {
		// Convert path to PascalCase name
		return toPascalCase(item.Path)
	}
	return toPascalCase(item.Name)
}

// generateRoutePath generates route path
func generateRoutePath(item *menusvc.MenuItem) string {
	if item.Path == "" {
		return ""
	}
	// For child routes (parentId != 0), return relative path without leading /
	// Vue Router will append this to parent path
	if item.ParentId != 0 {
		path := item.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		return path
	}
	// For root routes, ensure path starts with /
	if item.Path[0] != '/' {
		return "/" + item.Path
	}
	return item.Path
}

// generateComponentPath generates component path for Vben
func generateComponentPath(component string) string {
	if component == "" {
		return ""
	}
	// Vben expects component path like #/views/xxx/index.vue
	if component[0] == '#' {
		return component
	}
	return "#/views/" + component
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]byte, 0, len(s))
	upperNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '_' || c == '/' || c == ' ' {
			upperNext = true
			continue
		}
		if upperNext {
			if c >= 'a' && c <= 'z' {
				c = c - 32
			}
			upperNext = false
		}
		result = append(result, c)
	}
	return string(result)
}