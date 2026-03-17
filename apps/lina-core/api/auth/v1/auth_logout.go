package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Logout API

type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"认证管理" summary:"用户登出"`
}

type LogoutRes struct{}
