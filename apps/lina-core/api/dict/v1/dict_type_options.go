package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// TypeOptionsReq defines the request for querying dictionary type options.
type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"字典管理" summary:"获取全部字典类型选项" dc:"获取所有正常状态的字典类型列表，用于字典数据管理页面的类型选择下拉框" permission:"system:dict:query"`
}

// TypeOptionsRes defines the response for querying dictionary type options.
type TypeOptionsRes struct {
	List []*entity.SysDictType `json:"list" dc:"字典类型选项列表" eg:"[]"`
}
