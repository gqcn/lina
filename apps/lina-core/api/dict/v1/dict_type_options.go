package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictType Options API

type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"字典管理" summary:"获取全部字典类型选项"`
}

type TypeOptionsRes struct {
	List []*entity.SysDictType `json:"list" dc:"字典类型选项列表"`
}
