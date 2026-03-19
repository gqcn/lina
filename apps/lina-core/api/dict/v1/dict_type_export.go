package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Export API

type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"字典管理" summary:"导出字典类型" operLog:"4" dc:"导出字典类型数据为Excel文件，支持按条件筛选导出"`
	Name   string `json:"name" dc:"按字典名称筛选（模糊匹配）" eg:"性别"`
	Type   string `json:"type" dc:"按字典类型标识筛选（模糊匹配）" eg:"sys_user"`
}

type TypeExportRes struct{}
