package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Update API

type TypeUpdateReq struct {
	g.Meta `path:"/dict/type/{id}" method:"put" tags:"字典管理" summary:"更新字典类型"`
	Id     int     `json:"id" v:"required" dc:"字典类型ID"`
	Name   *string `json:"name" dc:"字典名称"`
	Type   *string `json:"type" dc:"字典类型"`
	Status *int    `json:"status" dc:"状态"`
	Remark *string `json:"remark" dc:"备注"`
}

type TypeUpdateRes struct{}
