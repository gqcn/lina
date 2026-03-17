package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Export API

type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"操作日志" summary:"导出操作日志" operLog:"4"`
	Title          string `json:"title" dc:"按模块标题筛选"`
	OperName       string `json:"operName" dc:"按操作人员筛选"`
	OperType       *int   `json:"operType" dc:"按操作类型筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按操作时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按操作时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向"`
}

type ExportRes struct{}
