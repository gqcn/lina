package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Delete API

type DataDeleteReq struct {
	g.Meta `path:"/dict/data/{id}" method:"delete" tags:"字典管理" summary:"删除字典数据" dc:"删除指定的字典数据项"`
	Id     int `json:"id" v:"required" dc:"字典数据ID" eg:"1"`
}

type DataDeleteRes struct{}
