package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dict Type API

type TypeListReq struct {
	g.Meta   `path:"/dict/type" method:"get" tags:"DictType" summary:"Get dict type list"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	Name     string `json:"name" dc:"Filter by dict name"`
	Type     string `json:"type" dc:"Filter by dict type"`
}

type TypeListRes struct {
	List  []*entity.SysDictType `json:"list" dc:"Dict type list"`
	Total int                   `json:"total" dc:"Total count"`
}

type TypeCreateReq struct {
	g.Meta `path:"/dict/type" method:"post" tags:"DictType" summary:"Create dict type"`
	Name   string `json:"name" v:"required#请输入字典名称" dc:"Dict name"`
	Type   string `json:"type" v:"required#请输入字典类型" dc:"Dict type"`
	Status *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled"`
	Remark string `json:"remark" dc:"Remark"`
}

type TypeCreateRes struct {
	Id int `json:"id" dc:"Dict type ID"`
}

type TypeGetReq struct {
	g.Meta `path:"/dict/type/{id}" method:"get" tags:"DictType" summary:"Get dict type detail"`
	Id     int `json:"id" v:"required" dc:"Dict type ID"`
}

type TypeGetRes struct {
	*entity.SysDictType `dc:"Dict type info"`
}

type TypeUpdateReq struct {
	g.Meta `path:"/dict/type/{id}" method:"put" tags:"DictType" summary:"Update dict type"`
	Id     int     `json:"id" v:"required" dc:"Dict type ID"`
	Name   *string `json:"name" dc:"Dict name"`
	Type   *string `json:"type" dc:"Dict type"`
	Status *int    `json:"status" dc:"Status"`
	Remark *string `json:"remark" dc:"Remark"`
}

type TypeUpdateRes struct{}

type TypeDeleteReq struct {
	g.Meta `path:"/dict/type/{id}" method:"delete" tags:"DictType" summary:"Delete dict type"`
	Id     int `json:"id" v:"required" dc:"Dict type ID"`
}

type TypeDeleteRes struct{}

type TypeExportReq struct {
	g.Meta `path:"/dict/type/export" method:"get" tags:"DictType" summary:"Export dict types to Excel" operLog:"4"`
	Name   string `json:"name" dc:"Filter by dict name"`
	Type   string `json:"type" dc:"Filter by dict type"`
}

type TypeExportRes struct{}

type TypeOptionsReq struct {
	g.Meta `path:"/dict/type/options" method:"get" tags:"DictType" summary:"Get all dict type options"`
}

type TypeOptionsRes struct {
	List []*entity.SysDictType `json:"list" dc:"Dict type options"`
}
