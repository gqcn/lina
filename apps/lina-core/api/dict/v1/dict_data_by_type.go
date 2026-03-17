package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictData ByType API

type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"字典管理" summary:"按类型获取字典数据"`
	DictType string `json:"dictType" v:"required" dc:"字典类型"`
}

type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"字典数据列表"`
}
