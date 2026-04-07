package v1

import "github.com/gogf/gf/v2/frame/g"

// RuntimeListReq is the request for querying public plugin runtime states.
type RuntimeListReq struct {
	g.Meta `path:"/plugins/runtime" method:"get" tags:"插件管理" summary:"查询插件运行状态" dc:"返回前端公共壳层渲染插件 Slot 所需的最小运行状态集合，供登录页和布局壳层在匿名或登录态下判断插件内容是否应显示"`
}

// RuntimeListRes is the response for querying public plugin runtime states.
type RuntimeListRes struct {
	List []*PluginRuntimeItem `json:"list" dc:"插件运行状态列表" eg:"[]"`
}

// PluginRuntimeItem represents public runtime state of one plugin.
type PluginRuntimeItem struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装/已集成 0=未安装" eg:"1"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
	StatusKey string `json:"statusKey" dc:"插件状态在系统插件注册表中的定位键名" eg:"sys_plugin.status:plugin-demo"`
}
