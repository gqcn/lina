package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User UpdateStatus API

type UpdateStatusReq struct {
	g.Meta `path:"/user/{id}/status" method:"put" tags:"用户管理" summary:"更新用户状态"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
	Status int `json:"status" v:"in:0,1#状态值无效" dc:"状态：1=正常 0=停用"`
}

type UpdateStatusRes struct{}
