package v1

import "github.com/gogf/gf/v2/frame/g"

// ReviewSummaryReq is the request for querying one review-friendly backend summary.
type ReviewSummaryReq struct {
	g.Meta `path:"/plugins/plugin-demo-dynamic/review-summary" method:"get" tags:"动态插件示例" summary:"查询动态插件 review 摘要" dc:"返回 plugin-demo-dynamic 后端目录中的 review 示例文案。当前迭代尚未支持动态插件后端动态执行，因此该接口定义仅用于展示动态插件未来如何组织后端 API 代码"`
}

// ReviewSummaryRes is the response for querying one review-friendly backend summary.
type ReviewSummaryRes struct {
	Message string `json:"message" dc:"用于人工 review 的动态插件后端示例说明" eg:"This backend example shows how a dynamic plugin can organize Go APIs before phase-3 runtime execution is implemented."`
}
