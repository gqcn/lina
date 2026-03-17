package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/operlog" method:"get" tags:"操作日志" summary:"获取操作日志列表"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Title          string `json:"title" dc:"按模块标题筛选"`
	OperName       string `json:"operName" dc:"按操作人员筛选"`
	OperType       *int   `json:"operType" dc:"按操作类型筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按操作时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按操作时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,oper_time,cost_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc"`
}

type ListRes struct {
	Items []*entity.SysOperLog `json:"items" dc:"操作日志列表"`
	Total int                  `json:"total" dc:"总条数"`
}

type GetReq struct {
	g.Meta `path:"/operlog/{id}" method:"get" tags:"操作日志" summary:"获取操作日志详情"`
	Id     int `json:"id" v:"required" dc:"操作日志ID"`
}

type GetRes struct {
	*entity.SysOperLog
}

type CleanReq struct {
	g.Meta    `path:"/operlog/clean" method:"delete" tags:"操作日志" summary:"清空操作日志"`
	BeginTime string `json:"beginTime" dc:"清理起始时间"`
	EndTime   string `json:"endTime" dc:"清理截止时间"`
}

type CleanRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}

type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"操作日志" summary:"导出操作日志" operLog:"4"`
	Title          string `json:"title" dc:"按模块标题筛选"`
	OperName       string `json:"operName" dc:"按操作人员筛选"`
	OperType       *int   `json:"operType" dc:"按操作类型筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按操作时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按操作时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向"`
}

type ExportRes struct{}

type DeleteReq struct {
	g.Meta `path:"/operlog/{ids}" method:"delete" tags:"操作日志" summary:"删除操作日志"`
	Ids    string `json:"ids" v:"required" dc:"日志ID，多个用逗号分隔"`
}

type DeleteRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}
