package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Export API

type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"用户管理" summary:"导出用户数据" operLog:"4"`
	Ids    []int `json:"ids" dc:"导出指定用户ID列表"`
}

type ExportRes struct{}
