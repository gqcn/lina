package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetAllReq defines the request for querying all user menu routes.
type GetAllReq struct {
	g.Meta `path:"/menus/all" method:"get" tags:"菜单管理" summary:"获取用户菜单路由" dc:"获取当前登录用户的菜单列表，返回Vben框架所需的路由格式，用于前端动态路由生成"`
}

// MenuRouteItem represents a menu route item for Vben frontend
type MenuRouteItem struct {
	Id        int              `json:"id" dc:"菜单ID" eg:"1"`
	ParentId  int              `json:"parentId" dc:"父菜单ID" eg:"0"`
	Name      string           `json:"name" dc:"路由名称（唯一）" eg:"System"`
	Path      string           `json:"path" dc:"路由路径" eg:"/system"`
	Component string           `json:"component" dc:"组件路径" eg:"#/views/system/user/index.vue"`
	Redirect  string           `json:"redirect,omitempty" dc:"重定向路径" eg:"/system/user"`
	Meta      *MenuRouteMeta   `json:"meta" dc:"路由元信息" eg:""`
	Children  []*MenuRouteItem `json:"children,omitempty" dc:"子路由列表" eg:"[]"`
}

// MenuRouteMeta represents route metadata for Vben
type MenuRouteMeta struct {
	Title            string `json:"title" dc:"菜单标题" eg:"系统管理"`
	Icon             string `json:"icon,omitempty" dc:"菜单图标" eg:"ant-design:setting-outlined"`
	ActiveIcon       string `json:"activeIcon,omitempty" dc:"激活时图标" eg:"ant-design:setting-filled"`
	HideInMenu       bool   `json:"hideInMenu,omitempty" dc:"是否在菜单中隐藏" eg:"false"`
	HideInBreadcrumb bool   `json:"hideInBreadcrumb,omitempty" dc:"是否在面包屑中隐藏" eg:"false"`
	HideInTab        bool   `json:"hideInTab,omitempty" dc:"是否在标签页中隐藏" eg:"false"`
	KeepAlive        bool   `json:"keepAlive,omitempty" dc:"是否缓存页面" eg:"true"`
	Order            int    `json:"order,omitempty" dc:"排序号" eg:"1"`
	Authority        string `json:"authority,omitempty" dc:"权限标识" eg:"system:user:list"`
	IgnoreAccess     bool   `json:"ignoreAccess,omitempty" dc:"是否忽略权限" eg:"false"`
}

// GetAllRes defines the wrapped response for user menu routes.
type GetAllRes struct {
	List []*MenuRouteItem `json:"list" dc:"用户菜单路由列表" eg:"[]"`
}

// GetAllResAlt defines the direct-array response for Vben compatibility.
type GetAllResAlt struct {
	g.Meta `path:"/menus/all" method:"get" tags:"菜单管理" summary:"获取用户菜单路由" dc:"获取当前登录用户的菜单列表，返回Vben框架所需的路由格式，用于前端动态路由生成"`
}

// MenuRouteList is a wrapper type for JSON array response
type MenuRouteList []*MenuRouteItem
