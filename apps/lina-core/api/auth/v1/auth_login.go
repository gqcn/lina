package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Login API

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"认证管理" summary:"用户登录"`
	Username string `json:"username" v:"required#请输入用户名"`
	Password string `json:"password" v:"required#请输入密码"`
}

type LoginRes struct {
	AccessToken string `json:"accessToken" dc:"JWT令牌"`
}
