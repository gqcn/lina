package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Delete API

type DeleteReq struct {
	g.Meta `path:"/dept/{id}" method:"delete" tags:"部门管理" summary:"删除部门"`
	Id     int `json:"id" v:"required" dc:"部门ID"`
}

type DeleteRes struct{}
