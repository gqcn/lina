package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Codes API

type CodesReq struct {
	g.Meta `path:"/auth/codes" method:"get" tags:"认证管理" summary:"获取权限码" dc:"获取当前登录用户的权限码列表，用于前端权限控制"`
}

type CodesRes struct {
	Codes []string `json:"codes" dc:"权限码列表" eg:"['user:list','user:create','user:update']"`
}
