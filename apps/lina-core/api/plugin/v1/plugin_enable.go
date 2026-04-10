package v1

import "github.com/gogf/gf/v2/frame/g"

// EnableReq is the request for enabling plugin.
type EnableReq struct {
	g.Meta `path:"/plugins/{id}/enable" method:"put" tags:"插件管理" summary:"启用插件" dc:"将指定插件标记为启用状态，并写入插件状态配置"`
	Id     string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
}

// EnableRes is the response for enabling plugin.
type EnableRes struct {
	Id      string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Enabled int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"1"`
}
