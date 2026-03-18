package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Create API

type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"用户管理" summary:"创建用户"`
	Username string `json:"username" v:"required|length:2,64#请输入用户名|用户名长度为2-64个字符"`
	Password string `json:"password" v:"required|length:6,32#请输入密码|密码长度为6-32个字符"`
	Nickname string `json:"nickname" v:"required#请输入昵称" dc:"昵称"`
	Email    string `json:"email" dc:"邮箱"`
	Phone    string `json:"phone" dc:"手机号"`
	Sex      *int   `json:"sex" d:"0" dc:"性别：0=未知 1=男 2=女"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark   string `json:"remark" dc:"备注"`
	DeptId   *int   `json:"deptId" dc:"部门ID"`
	PostIds  []int  `json:"postIds" dc:"岗位ID列表"`
}

type CreateRes struct {
	Id int `json:"id" dc:"用户ID"`
}
