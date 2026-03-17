package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/loginlog" method:"get" tags:"登录日志" summary:"获取登录日志列表"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	UserName       string `json:"userName" dc:"按用户名筛选"`
	Ip             string `json:"ip" dc:"按IP地址筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按登录时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按登录时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,login_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc"`
}

type ListRes struct {
	Items []*entity.SysLoginLog `json:"items" dc:"登录日志列表"`
	Total int                   `json:"total" dc:"总条数"`
}

type GetReq struct {
	g.Meta `path:"/loginlog/{id}" method:"get" tags:"登录日志" summary:"获取登录日志详情"`
	Id     int `json:"id" v:"required" dc:"登录日志ID"`
}

type GetRes struct {
	*entity.SysLoginLog
}

type CleanReq struct {
	g.Meta    `path:"/loginlog/clean" method:"delete" tags:"登录日志" summary:"清空登录日志"`
	BeginTime string `json:"beginTime" dc:"清理起始时间"`
	EndTime   string `json:"endTime" dc:"清理截止时间"`
}

type CleanRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}

type ExportReq struct {
	g.Meta         `path:"/loginlog/export" method:"get" tags:"登录日志" summary:"导出登录日志" operLog:"4"`
	UserName       string `json:"userName" dc:"按用户名筛选"`
	Ip             string `json:"ip" dc:"按IP地址筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按登录时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按登录时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向"`
}

type ExportRes struct{}

type DeleteReq struct {
	g.Meta `path:"/loginlog/{ids}" method:"delete" tags:"登录日志" summary:"删除登录日志"`
	Ids    string `json:"ids" v:"required" dc:"日志ID，多个用逗号分隔"`
}

type DeleteRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}
