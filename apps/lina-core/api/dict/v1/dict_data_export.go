package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Export API

type DataExportReq struct {
	g.Meta   `path:"/dict/data/export" method:"get" tags:"字典管理" summary:"导出字典数据" operLog:"4"`
	DictType string `json:"dictType" dc:"按字典类型筛选"`
	Label    string `json:"label" dc:"按标签筛选"`
}

type DataExportRes struct{}
