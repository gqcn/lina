package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Update API

type UpdateReq struct {
	g.Meta `path:"/post/{id}" method:"put" tags:"岗位管理" summary:"更新岗位"`
	Id     int     `json:"id" v:"required" dc:"岗位ID"`
	DeptId *int    `json:"deptId" dc:"部门ID"`
	Code   *string `json:"code" dc:"岗位编码"`
	Name   *string `json:"name" dc:"岗位名称"`
	Sort   *int    `json:"sort" dc:"排序号"`
	Status *int    `json:"status" dc:"状态"`
	Remark *string `json:"remark" dc:"备注"`
}

type UpdateRes struct{}
