package v1

import "github.com/gogf/gf/v2/frame/g"

// SummaryReq is the request for querying plugin-demo summary.
type SummaryReq struct {
	g.Meta `path:"/plugins/plugin-demo/summary" method:"get" tags:"插件示例" summary:"查询插件示例摘要" dc:"返回 plugin-demo 页面展示所需的简要介绍文案，用于验证插件页面可读取插件后端接口数据"`
}

// SummaryRes is the response for querying plugin-demo summary.
type SummaryRes struct {
	Message string `json:"message" dc:"页面展示使用的简要介绍文案，来自插件后端接口" eg:"这是一条来自 plugin-demo 接口的简要介绍，用于验证插件页面可读取插件后端数据。"`
}
