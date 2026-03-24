package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Config Export API

// ExportReq defines the request for exporting configs to Excel.
type ExportReq struct {
	g.Meta    `path:"/config/export" method:"get" tags:"参数设置" summary:"导出参数设置" operLog:"4" dc:"导出参数设置数据为Excel文件，支持按条件筛选导出"`
	Name      string `json:"name" dc:"按参数名称筛选（模糊匹配）" eg:"主框架页"`
	Key       string `json:"key" dc:"按参数键名筛选（模糊匹配）" eg:"sys.index"`
	BeginTime string `json:"beginTime" dc:"创建时间范围-开始时间，格式YYYY-MM-DD" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"创建时间范围-结束时间，格式YYYY-MM-DD" eg:"2025-12-31"`
}

// ExportRes 参数设置导出响应
type ExportRes struct{}
