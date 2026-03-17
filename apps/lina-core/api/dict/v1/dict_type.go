package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dict Type API

type TypeListReq struct {
	g.Meta   `path:"/dict/type" method:"get" tags:"字典管理" summary:"获取字典类型列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Name     string `json:"name" dc:"按字典名称筛选"`
	Type     string `json:"type" dc:"按字典类型筛选"`
}

type TypeListRes struct {
	List  []*entity.SysDictType `json:"list" dc:"字典类型列表"`
	Total int                   `json:"total" dc:"总条数"`
}

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

type TypeGetReq struct {
	g.Meta `path:"/dict/type/{id}" method:"get" tags:"字典管理" summary:"获取字典类型详情"`
	Id     int `json:"id" v:"required" dc:"字典类型ID"`
}

type TypeGetRes struct {
	*entity.SysDictType `dc:"字典类型信息"`
}

type TypeUpdateReq struct {
	g.Meta `path:"/dict/type/{id}" method:"put" tags:"字典管理" summary:"更新字典类型"`
	Id     int     `json:"id" v:"required" dc:"字典类型ID"`
	Name   *string `json:"name" dc:"字典名称"`
	Type   *string `json:"type" dc:"字典类型"`
	Status *int    `json:"status" dc:"状态"`
	Remark *string `json:"remark" dc:"备注"`
}

type TypeUpdateRes struct{}

type TypeDeleteReq struct {
	g.Meta `path:"/dict/type/{id}" method:"delete" tags:"字典管理" summary:"删除字典类型"`
	Id     int `json:"id" v:"required" dc:"字典类型ID"`
}

type TypeDeleteRes struct{}

type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"字典管理" summary:"导出字典类型" operLog:"4"`
	Name   string `json:"name" dc:"按字典名称筛选"`
	Type   string `json:"type" dc:"按字典类型筛选"`
}

type TypeExportRes struct{}

type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"字典管理" summary:"获取全部字典类型选项"`
}

type TypeOptionsRes struct {
	List []*entity.SysDictType `json:"list" dc:"字典类型选项列表"`
}
