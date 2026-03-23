package v1

import "github.com/gogf/gf/v2/frame/g"

type JobLogListReq struct {
	g.Meta    `path:"/job/log/list" method:"get" tags:"定时任务" summary:"查询执行日志" dc:"分页查询任务执行日志，支持按任务名称、状态、时间范围筛选"`
	JobName   string `json:"jobName" dc:"任务名称，支持模糊查询" eg:"会话清理"`
	Status    *int   `json:"status" dc:"执行状态：1=成功 0=失败，不传则查询全部" eg:"1"`
	StartTime string `json:"startTime" dc:"开始时间(起)" eg:"2026-03-01 00:00:00"`
	EndTime   string `json:"endTime" dc:"开始时间(止)" eg:"2026-03-31 23:59:59"`
	Page      int    `json:"page" v:"required|min:1" dc:"页码" eg:"1"`
	PageSize  int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量" eg:"10"`
}

type JobLogListRes struct {
	Items []*JobLogItem `json:"items" dc:"日志列表"`
	Total int           `json:"total" dc:"总数"`
}

type JobLogItem struct {
	Id         uint64 `json:"id" dc:"日志ID"`
	JobId      uint64 `json:"jobId" dc:"任务ID"`
	JobName    string `json:"jobName" dc:"任务名称"`
	JobGroup   string `json:"jobGroup" dc:"任务分组"`
	Command    string `json:"command" dc:"执行指令"`
	Status     int    `json:"status" dc:"执行状态：1=成功 0=失败"`
	StartTime  string `json:"startTime" dc:"开始时间"`
	EndTime    string `json:"endTime" dc:"结束时间"`
	Duration   int    `json:"duration" dc:"执行耗时(毫秒)"`
	ErrorMsg   string `json:"errorMsg" dc:"错误信息"`
	CreateTime string `json:"createTime" dc:"创建时间"`
}
