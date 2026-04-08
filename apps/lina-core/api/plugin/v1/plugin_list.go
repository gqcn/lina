package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq is the request for querying plugin list.
type ListReq struct {
	g.Meta    `path:"/plugins" method:"get" tags:"插件管理" summary:"查询插件列表" dc:"扫描源码插件目录并同步插件基础状态，返回插件清单与启用状态"`
	Id        string `json:"id" dc:"按插件唯一标识筛选，模糊匹配，不传则查询全部" eg:"plugin-demo"`
	Name      string `json:"name" dc:"按插件名称筛选，模糊匹配，不传则查询全部" eg:"示例插件"`
	Type      string `json:"type" dc:"按插件类型筛选：source=源码插件 runtime=运行时插件（同时匹配package/wasm） package=打包运行时插件 wasm=WASM运行时插件，不传则查询全部" eg:"runtime"`
	Status    *int   `json:"status" dc:"按启用状态筛选：1=启用 0=禁用，不传则查询全部" eg:"1"`
	Installed *int   `json:"installed" dc:"按安装状态筛选：1=已安装/已集成 0=未安装，不传则查询全部；源码插件默认视为已集成" eg:"1"`
}

// ListRes is the response for querying plugin list.
type ListRes struct {
	List  []*PluginItem `json:"list" dc:"插件列表" eg:"[]"`
	Total int           `json:"total" dc:"插件总数" eg:"1"`
}

// PluginItem represents plugin information.
type PluginItem struct {
	Id          string `json:"id" dc:"插件唯一标识" eg:"plugin-demo"`
	Name        string `json:"name" dc:"插件名称" eg:"示例插件"`
	Version     string `json:"version" dc:"插件版本号" eg:"0.1.0"`
	Type        string `json:"type" dc:"插件交付类型：source=源码插件 package=运行时插件(package) wasm=运行时插件(wasm)" eg:"source"`
	Runtime     string `json:"runtime" dc:"交付/运行时方式：source=源码集成 package=打包发布 wasm=WASM发布" eg:"source"`
	Entry       string `json:"entry" dc:"插件入口描述，可为路由、文件或模块标识" eg:"backend:plugins/{id}/summary"`
	Description string `json:"description" dc:"插件描述" eg:"提供左侧菜单页面、前端 Slot 与公开/受保护路由示例的源码插件"`
	Installed   int    `json:"installed" dc:"安装状态：1=已安装/已集成 0=未安装；源码插件默认返回1表示已随宿主集成" eg:"1"`
	InstalledAt string `json:"installedAt" dc:"插件安装或源码接入时间，未安装时返回空字符串" eg:"2026-01-01 12:00:00"`
	Enabled     int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
	StatusKey   string `json:"statusKey" dc:"插件状态在系统插件注册表中的定位键名" eg:"sys_plugin.status:plugin-demo"`
	UpdatedAt   string `json:"updatedAt" dc:"插件注册表最后更新时间" eg:"2026-01-01 12:00:00"`
}
