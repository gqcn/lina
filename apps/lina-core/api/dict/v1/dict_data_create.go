package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictData Create API

type DataCreateReq struct {
	g.Meta   `path:"/dict/data" method:"post" tags:"字典管理" summary:"创建字典数据"`
	DictType string `json:"dictType" v:"required#请输入字典类型" dc:"字典类型"`
	Label    string `json:"label" v:"required#请输入字典标签" dc:"标签"`
	Value    string `json:"value" v:"required#请输入字典值" dc:"值"`
	Sort     *int   `json:"sort" d:"0" dc:"排序号"`
	TagStyle string `json:"tagStyle" dc:"标签样式"`
	CssClass string `json:"cssClass" dc:"CSS类名"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark   string `json:"remark" dc:"备注"`
}

type DataCreateRes struct {
	Id int `json:"id" dc:"字典数据ID"`
}
