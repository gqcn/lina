package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Update API

type DataUpdateReq struct {
	g.Meta   `path:"/dict/data/{id}" method:"put" tags:"字典管理" summary:"更新字典数据"`
	Id       int     `json:"id" v:"required" dc:"字典数据ID"`
	DictType *string `json:"dictType" dc:"字典类型"`
	Label    *string `json:"label" dc:"标签"`
	Value    *string `json:"value" dc:"值"`
	Sort     *int    `json:"sort" dc:"排序号"`
	TagStyle *string `json:"tagStyle" dc:"标签样式"`
	CssClass *string `json:"cssClass" dc:"CSS类名"`
	Status   *int    `json:"status" dc:"状态"`
	Remark   *string `json:"remark" dc:"备注"`
}

type DataUpdateRes struct{}
