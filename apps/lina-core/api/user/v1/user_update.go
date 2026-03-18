package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Update API

type UpdateReq struct {
	g.Meta   `path:"/user/{id}" method:"put" tags:"用户管理" summary:"更新用户"`
	Id       int     `json:"id" v:"required" dc:"用户ID"`
	Username *string `json:"username" dc:"用户名"`
	Password *string `json:"password" dc:"密码（为空则不修改）"`
	Nickname *string `json:"nickname" v:"required#请输入昵称" dc:"昵称"`
	Email    *string `json:"email" dc:"邮箱"`
	Phone    *string `json:"phone" dc:"手机号"`
	Sex      *int    `json:"sex" dc:"性别"`
	Status   *int    `json:"status" dc:"状态"`
	Remark   *string `json:"remark" dc:"备注"`
	DeptId   *int    `json:"deptId" dc:"部门ID"`
	PostIds  []int   `json:"postIds" dc:"岗位ID列表"`
}

type UpdateRes struct{}
