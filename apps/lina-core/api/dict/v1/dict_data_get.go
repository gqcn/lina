package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictData Get API

type DataGetReq struct {
	g.Meta `path:"/dict/data/{id}" method:"get" tags:"字典管理" summary:"获取字典数据详情"`
	Id     int `json:"id" v:"required" dc:"字典数据ID"`
}

type DataGetRes struct {
	*entity.SysDictData `dc:"字典数据信息"`
}
