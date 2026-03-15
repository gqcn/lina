package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/loginlog" method:"get" tags:"LoginLog" summary:"Get login log list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	UserName       string `json:"userName" dc:"Filter by username"`
	Ip             string `json:"ip" dc:"Filter by IP"`
	Status         *int   `json:"status" dc:"Filter by status"`
	BeginTime      string `json:"beginTime" dc:"Filter by login_time start"`
	EndTime        string `json:"endTime" dc:"Filter by login_time end"`
	OrderBy        string `json:"orderBy" dc:"Sort field: id,login_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc"`
}

type ListRes struct {
	Items []*entity.SysLoginLog `json:"items" dc:"Login log list"`
	Total int                   `json:"total" dc:"Total count"`
}

type GetReq struct {
	g.Meta `path:"/loginlog/{id}" method:"get" tags:"LoginLog" summary:"Get login log detail"`
	Id     int `json:"id" v:"required" dc:"Login log ID"`
}

type GetRes struct {
	*entity.SysLoginLog
}

type CleanReq struct {
	g.Meta    `path:"/loginlog/clean" method:"delete" tags:"LoginLog" summary:"Clean login logs"`
	BeginTime string `json:"beginTime" dc:"Clean logs from this time"`
	EndTime   string `json:"endTime" dc:"Clean logs until this time"`
}

type CleanRes struct {
	Deleted int `json:"deleted" dc:"Number of deleted records"`
}

type ExportReq struct {
	g.Meta         `path:"/loginlog/export" method:"get" tags:"LoginLog" summary:"Export login logs" operLog:"4"`
	UserName       string `json:"userName" dc:"Filter by username"`
	Ip             string `json:"ip" dc:"Filter by IP"`
	Status         *int   `json:"status" dc:"Filter by status"`
	BeginTime      string `json:"beginTime" dc:"Filter by login_time start"`
	EndTime        string `json:"endTime" dc:"Filter by login_time end"`
	OrderBy        string `json:"orderBy" dc:"Sort field"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction"`
}

type ExportRes struct{}

type DeleteReq struct {
	g.Meta `path:"/loginlog/{ids}" method:"delete" tags:"LoginLog" summary:"Delete login logs"`
	Ids    string `json:"ids" v:"required" dc:"Comma-separated log IDs"`
}

type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of deleted records"`
}
