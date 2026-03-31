package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Menu Get API - get menu detail by ID

type GetReq struct {
	g.Meta `path:"/menu/:id" method:"get" tags:"菜单管理" summary:"获取菜单详情" dc:"根据菜单ID获取菜单详情信息，包含父菜单名称"`
	Id     int `json:"id" v:"required|min:1" dc:"菜单ID" eg:"1"`
}

type GetRes struct {
	*MenuItem
	ParentName string `json:"parentName" dc:"父菜单名称" eg:"系统管理"`
}