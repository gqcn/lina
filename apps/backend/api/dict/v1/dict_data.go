package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dict Data API

type DataListReq struct {
	g.Meta   `path:"/dict/data" method:"get" tags:"DictData" summary:"Get dict data list"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	DictType string `json:"dictType" dc:"Filter by dict type"`
	Label    string `json:"label" dc:"Filter by label"`
}

type DataListRes struct {
	List  []*entity.SysDictData `json:"list" dc:"Dict data list"`
	Total int                   `json:"total" dc:"Total count"`
}

type DataCreateReq struct {
	g.Meta   `path:"/dict/data" method:"post" tags:"DictData" summary:"Create dict data"`
	DictType string `json:"dictType" v:"required#请输入字典类型" dc:"Dict type"`
	Label    string `json:"label" v:"required#请输入字典标签" dc:"Label"`
	Value    string `json:"value" v:"required#请输入字典值" dc:"Value"`
	Sort     *int   `json:"sort" d:"0" dc:"Sort order"`
	TagStyle string `json:"tagStyle" dc:"Tag style"`
	CssClass string `json:"cssClass" dc:"CSS class"`
	Status   *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled"`
	Remark   string `json:"remark" dc:"Remark"`
}

type DataCreateRes struct {
	Id int `json:"id" dc:"Dict data ID"`
}

type DataGetReq struct {
	g.Meta `path:"/dict/data/{id}" method:"get" tags:"DictData" summary:"Get dict data detail"`
	Id     int `json:"id" v:"required" dc:"Dict data ID"`
}

type DataGetRes struct {
	*entity.SysDictData `dc:"Dict data info"`
}

type DataUpdateReq struct {
	g.Meta   `path:"/dict/data/{id}" method:"put" tags:"DictData" summary:"Update dict data"`
	Id       int     `json:"id" v:"required" dc:"Dict data ID"`
	DictType *string `json:"dictType" dc:"Dict type"`
	Label    *string `json:"label" dc:"Label"`
	Value    *string `json:"value" dc:"Value"`
	Sort     *int    `json:"sort" dc:"Sort order"`
	TagStyle *string `json:"tagStyle" dc:"Tag style"`
	CssClass *string `json:"cssClass" dc:"CSS class"`
	Status   *int    `json:"status" dc:"Status"`
	Remark   *string `json:"remark" dc:"Remark"`
}

type DataUpdateRes struct{}

type DataDeleteReq struct {
	g.Meta `path:"/dict/data/{id}" method:"delete" tags:"DictData" summary:"Delete dict data"`
	Id     int `json:"id" v:"required" dc:"Dict data ID"`
}

type DataDeleteRes struct{}

type DataExportReq struct {
	g.Meta   `path:"/dict/data/export" method:"get" tags:"DictData" summary:"Export dict data to Excel"`
	DictType string `json:"dictType" dc:"Filter by dict type"`
	Label    string `json:"label" dc:"Filter by label"`
}

type DataExportRes struct{}

type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"DictData" summary:"Get dict data by type"`
	DictType string `json:"dictType" v:"required" dc:"Dict type"`
}

type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"Dict data list"`
}
