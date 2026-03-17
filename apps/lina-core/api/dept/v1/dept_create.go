package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Create API

type CreateReq struct {
	g.Meta   `path:"/dept" method:"post" tags:"部门管理" summary:"创建部门"`
	ParentId int    `json:"parentId" d:"0" dc:"父级部门ID"`
	Name     string `json:"name" v:"required#请输入部门名称" dc:"部门名称"`
	Code     string `json:"code" dc:"部门编码（唯一）"`
	OrderNum *int   `json:"orderNum" d:"0" dc:"排序号"`
	Leader   *int   `json:"leader" dc:"负责人用户ID"`
	Phone    string `json:"phone" dc:"联系电话"`
	Email    string `json:"email" dc:"邮箱"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark   string `json:"remark" dc:"备注"`
}

type CreateRes struct {
	Id int `json:"id" dc:"部门ID"`
}
