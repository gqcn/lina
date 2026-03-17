package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Info API

type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"用户管理" summary:"获取前端用户信息"`
}

type GetInfoRes struct {
	UserId   int      `json:"userId" dc:"用户ID"`
	Username string   `json:"username" dc:"用户名"`
	RealName string   `json:"realName" dc:"真实姓名"`
	Avatar   string   `json:"avatar" dc:"头像地址"`
	Roles    []string `json:"roles" dc:"用户角色"`
	HomePath string   `json:"homePath" dc:"首页路径"`
}
