package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"Auth" summary:"User login"`
	Username string `json:"username" v:"required#请输入用户名"`
	Password string `json:"password" v:"required#请输入密码"`
}

type LoginRes struct {
	AccessToken string `json:"accessToken" dc:"JWT Token"`
}

type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"Auth" summary:"User logout"`
}

type LogoutRes struct{}

type CodesReq struct {
	g.Meta `path:"/auth/codes" method:"get" tags:"Auth" summary:"Get access codes"`
}

type CodesRes struct {
	Codes []string `json:"codes" dc:"Access codes"`
}

