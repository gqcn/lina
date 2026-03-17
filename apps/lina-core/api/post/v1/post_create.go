package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Create API

type CreateReq struct {
	g.Meta `path:"/post" method:"post" tags:"岗位管理" summary:"创建岗位"`
	DeptId int    `json:"deptId" v:"required#请选择所属部门" dc:"部门ID"`
	Code   string `json:"code" v:"required#请输入岗位编码" dc:"岗位编码"`
	Name   string `json:"name" v:"required#请输入岗位名称" dc:"岗位名称"`
	Sort   *int   `json:"sort" d:"0" dc:"排序号"`
	Status *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark string `json:"remark" dc:"备注"`
}

type CreateRes struct {
	Id int `json:"id" dc:"岗位ID"`
}
