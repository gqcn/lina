package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Update API

type UpdateReq struct {
	g.Meta   `path:"/dept/{id}" method:"put" tags:"部门管理" summary:"更新部门"`
	Id       int     `json:"id" v:"required" dc:"部门ID"`
	ParentId *int    `json:"parentId" dc:"父级部门ID"`
	Name     *string `json:"name" dc:"部门名称"`
	Code     *string `json:"code" dc:"部门编码（唯一）"`
	OrderNum *int    `json:"orderNum" dc:"排序号"`
	Leader   *int    `json:"leader" dc:"负责人用户ID"`
	Phone    *string `json:"phone" dc:"联系电话"`
	Email    *string `json:"email" dc:"邮箱"`
	Status   *int    `json:"status" dc:"状态"`
	Remark   *string `json:"remark" dc:"备注"`
}

type UpdateRes struct{}
