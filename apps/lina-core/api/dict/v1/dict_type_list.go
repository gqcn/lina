package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictType List API

type TypeListReq struct {
	g.Meta   `path:"/dict/type" method:"get" tags:"字典管理" summary:"获取字典类型列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Name     string `json:"name" dc:"按字典名称筛选"`
	Type     string `json:"type" dc:"按字典类型筛选"`
}

type TypeListRes struct {
	List  []*entity.SysDictType `json:"list" dc:"字典类型列表"`
	Total int                   `json:"total" dc:"总条数"`
}
