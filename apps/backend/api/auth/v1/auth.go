package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"认证管理" summary:"用户登录"`
	Username string `json:"username" v:"required#请输入用户名"`
	Password string `json:"password" v:"required#请输入密码"`
}

type LoginRes struct {
	AccessToken string `json:"accessToken" dc:"JWT令牌"`
}

type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"认证管理" summary:"用户登出"`
}

type LogoutRes struct{}

type CodesReq struct {
	g.Meta `path:"/auth/codes" method:"get" tags:"认证管理" summary:"获取权限码"`
}

type CodesRes struct {
	Codes []string `json:"codes" dc:"权限码列表"`
}
