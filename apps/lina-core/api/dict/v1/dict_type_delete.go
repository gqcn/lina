package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Delete API

type TypeDeleteReq struct {
	g.Meta `path:"/dict/type/{id}" method:"delete" tags:"字典管理" summary:"删除字典类型"`
	Id     int `json:"id" v:"required" dc:"字典类型ID"`
}

type TypeDeleteRes struct{}
