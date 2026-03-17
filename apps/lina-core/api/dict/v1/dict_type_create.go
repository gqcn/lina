package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DictType Create API

type TypeCreateReq struct {
	g.Meta `path:"/dict/type" method:"post" tags:"字典管理" summary:"创建字典类型"`
	Name   string `json:"name" v:"required#请输入字典名称" dc:"字典名称"`
	Type   string `json:"type" v:"required#请输入字典类型" dc:"字典类型"`
	Status *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark string `json:"remark" dc:"备注"`
}

type TypeCreateRes struct {
	Id int `json:"id" dc:"字典类型ID"`
}
