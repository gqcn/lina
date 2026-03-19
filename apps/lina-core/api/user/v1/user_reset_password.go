package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User ResetPassword API

type ResetPasswordReq struct {
	g.Meta   `path:"/user/{id}/reset-password" method:"put" tags:"用户管理" summary:"重置用户密码" dc:"管理员重置指定用户的登录密码"`
	Id       int    `json:"id" v:"required" dc:"用户ID" eg:"1"`
	Password string `json:"password" v:"required|length:5,20#请输入密码|密码长度为5-20个字符" dc:"新密码" eg:"123456"`
}

type ResetPasswordRes struct{}
