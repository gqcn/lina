package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Users API

// UsersReq returns users belonging to a dept (for leader selection).
// When Id=0, returns all users. When Id>0, returns users in the dept and all its sub-depts.
type UsersReq struct {
	g.Meta  `path:"/dept/{id}/users" method:"get" tags:"部门管理" summary:"获取部门用户列表"`
	Id      int    `json:"id" dc:"部门ID，0表示所有用户"`
	Keyword string `json:"keyword" dc:"按用户名或昵称搜索"`
	Limit   int    `json:"limit" d:"10" dc:"最大返回条数"`
}

type DeptUser struct {
	Id       int    `json:"id" dc:"用户ID"`
	Username string `json:"username" dc:"用户名"`
	Nickname string `json:"nickname" dc:"昵称"`
}

type UsersRes struct {
	List []*DeptUser `json:"list" dc:"用户列表"`
}
