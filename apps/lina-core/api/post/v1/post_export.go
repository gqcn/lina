package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Export API

type ExportReq struct {
	g.Meta `path:"/post/export" method:"get" tags:"岗位管理" summary:"导出岗位数据" operLog:"4"`
	DeptId *int   `json:"deptId" dc:"按部门ID筛选"`
	Code   string `json:"code" dc:"按岗位编码筛选"`
	Name   string `json:"name" dc:"按岗位名称筛选"`
	Status *int   `json:"status" dc:"按状态筛选"`
}

type ExportRes struct{}
