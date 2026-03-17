package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dict Data API

type DataListReq struct {
	g.Meta   `path:"/dict/data" method:"get" tags:"字典管理" summary:"获取字典数据列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	DictType string `json:"dictType" dc:"按字典类型筛选"`
	Label    string `json:"label" dc:"按标签筛选"`
}

type DataListRes struct {
	List  []*entity.SysDictData `json:"list" dc:"字典数据列表"`
	Total int                   `json:"total" dc:"总条数"`
}

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

type DataGetReq struct {
	g.Meta `path:"/dict/data/{id}" method:"get" tags:"字典管理" summary:"获取字典数据详情"`
	Id     int `json:"id" v:"required" dc:"字典数据ID"`
}

type DataGetRes struct {
	*entity.SysDictData `dc:"字典数据信息"`
}

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

type DataDeleteReq struct {
	g.Meta `path:"/dict/data/{id}" method:"delete" tags:"字典管理" summary:"删除字典数据"`
	Id     int `json:"id" v:"required" dc:"字典数据ID"`
}

type DataDeleteRes struct{}

type DataExportReq struct {
	g.Meta   `path:"/dict/data/export" method:"get" tags:"字典管理" summary:"导出字典数据" operLog:"4"`
	DictType string `json:"dictType" dc:"按字典类型筛选"`
	Label    string `json:"label" dc:"按标签筛选"`
}

type DataExportRes struct{}

type DataByTypeReq struct {
	g.Meta   `path:"/dict/data/type/{dictType}" method:"get" tags:"字典管理" summary:"按类型获取字典数据"`
	DictType string `json:"dictType" v:"required" dc:"字典类型"`
}

type DataByTypeRes struct {
	List []*entity.SysDictData `json:"list" dc:"字典数据列表"`
}
