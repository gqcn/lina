package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictData List API

type DataListReq struct {
	g.Meta   `path:"/dict/data" method:"get" tags:"字典管理" summary:"获取字典数据列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	DictType string `json:"dictType" dc:"按字典类型筛选"`
	Label    string `json:"label" dc:"按标签筛选"`
}

type DataListRes struct {
	List  []*entity.SysDictData `json:"list" dc:"字典数据列表"`
	Total int                   `json:"total" dc:"总条数"`
}
