package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictType Get API

type TypeGetReq struct {
	g.Meta `path:"/dict/type/{id}" method:"get" tags:"字典管理" summary:"获取字典类型详情"`
	Id     int `json:"id" v:"required" dc:"字典类型ID"`
}

type TypeGetRes struct {
	*entity.SysDictType `dc:"字典类型信息"`
}
