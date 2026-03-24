package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DictData ByType API

type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"字典管理" summary:"按类型获取字典数据" dc:"根据字典类型标识获取该类型下所有正常状态的字典数据项，用于前端下拉选项等场景"`
	DictType string `json:"dictType" v:"required" dc:"字典类型标识" eg:"sys_user_sex"`
}

// DataByTypeRes dictionary data by type response
type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"字典数据列表" eg:"[]"`
}
