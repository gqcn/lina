package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Export API

type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"字典管理" summary:"导出字典类型" operLog:"4"`
	Name   string `json:"name" dc:"按字典名称筛选"`
	Type   string `json:"type" dc:"按字典类型筛选"`
}

type TypeExportRes struct{}
