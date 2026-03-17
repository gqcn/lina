package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Delete API

type DeleteReq struct {
	g.Meta `path:"/user/{id}" method:"delete" tags:"用户管理" summary:"删除用户"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
}

type DeleteRes struct{}
