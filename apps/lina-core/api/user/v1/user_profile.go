package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// User Profile API

type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"用户管理" summary:"获取当前用户信息"`
}

type GetProfileRes struct {
	*entity.SysUser `dc:"用户信息"`
}

type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"用户管理" summary:"更新当前用户信息"`
	Nickname *string `json:"nickname" v:"required#请输入昵称" dc:"昵称"`
	Email    *string `json:"email" dc:"邮箱"`
	Phone    *string `json:"phone" dc:"手机号"`
	Sex      *int    `json:"sex" dc:"性别"`
	Password *string `json:"password" dc:"新密码"`
}

type UpdateProfileRes struct{}
