package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Info API

type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"用户管理" summary:"获取前端用户信息" dc:"获取当前登录用户的基本信息，包括用户名、角色、头像等，用于前端页面展示和权限控制"`
}

type GetInfoRes struct {
	UserId   int      `json:"userId" dc:"用户ID" eg:"1"`
	Username string   `json:"username" dc:"用户名" eg:"admin"`
	RealName string   `json:"realName" dc:"真实姓名（昵称）" eg:"管理员"`
	Email    string   `json:"email" dc:"邮箱地址" eg:"admin@example.com"`
	Avatar   string   `json:"avatar" dc:"头像地址" eg:"/upload/avatar/default.png"`
	Roles    []string `json:"roles" dc:"用户角色" eg:"['admin']"`
	HomePath string   `json:"homePath" dc:"首页路径" eg:"/dashboard"`
}
