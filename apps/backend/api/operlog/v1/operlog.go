package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/operlog" method:"get" tags:"OperLog" summary:"Get operation log list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	Title          string `json:"title" dc:"Filter by module title"`
	OperName       string `json:"operName" dc:"Filter by operator name"`
	OperType       *int   `json:"operType" dc:"Filter by operation type"`
	Status         *int   `json:"status" dc:"Filter by status"`
	BeginTime      string `json:"beginTime" dc:"Filter by oper_time start"`
	EndTime        string `json:"endTime" dc:"Filter by oper_time end"`
	OrderBy        string `json:"orderBy" dc:"Sort field: id,oper_time,cost_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc"`
}

type ListRes struct {
	Items []*entity.SysOperLog `json:"items" dc:"Operation log list"`
	Total int                  `json:"total" dc:"Total count"`
}

type GetReq struct {
	g.Meta `path:"/operlog/{id}" method:"get" tags:"OperLog" summary:"Get operation log detail"`
	Id     int `json:"id" v:"required" dc:"Operation log ID"`
}

type GetRes struct {
	*entity.SysOperLog
}

type CleanReq struct {
	g.Meta    `path:"/operlog/clean" method:"delete" tags:"OperLog" summary:"Clean operation logs"`
	BeginTime string `json:"beginTime" dc:"Clean logs from this time"`
	EndTime   string `json:"endTime" dc:"Clean logs until this time"`
}

type CleanRes struct {
	Deleted int `json:"deleted" dc:"Number of deleted records"`
}

type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"OperLog" summary:"Export operation logs" operLog:"4"`
	Title          string `json:"title" dc:"Filter by module title"`
	OperName       string `json:"operName" dc:"Filter by operator name"`
	OperType       *int   `json:"operType" dc:"Filter by operation type"`
	Status         *int   `json:"status" dc:"Filter by status"`
	BeginTime      string `json:"beginTime" dc:"Filter by oper_time start"`
	EndTime        string `json:"endTime" dc:"Filter by oper_time end"`
	OrderBy        string `json:"orderBy" dc:"Sort field"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction"`
}

type ExportRes struct{}

type DeleteReq struct {
	g.Meta `path:"/operlog/{ids}" method:"delete" tags:"OperLog" summary:"Delete operation logs"`
	Ids    string `json:"ids" v:"required" dc:"Comma-separated log IDs"`
}

type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of deleted records"`
}
