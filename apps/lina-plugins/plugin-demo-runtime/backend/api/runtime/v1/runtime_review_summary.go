package v1

import "github.com/gogf/gf/v2/frame/g"

// ReviewSummaryReq is the request for querying one review-friendly backend summary.
type ReviewSummaryReq struct {
	g.Meta `path:"/plugins/plugin-demo-runtime/review-summary" method:"get" tags:"运行时插件示例" summary:"查询运行时插件 review 摘要" dc:"返回 plugin-demo-runtime 后端目录中的 review 示例文案。当前迭代尚未支持 runtime 插件后端动态执行，因此该接口定义仅用于展示运行时插件未来如何组织后端 API 代码"`
}

// ReviewSummaryRes is the response for querying one review-friendly backend summary.
type ReviewSummaryRes struct {
	Message string `json:"message" dc:"用于人工 review 的运行时插件后端示例说明" eg:"This backend example shows how a runtime plugin can organize Go APIs before phase-3 runtime execution is implemented."`
}
