package v1

import "github.com/gogf/gf/v2/frame/g"

type JobListReq struct {
	g.Meta   `path:"/job/list" method:"get" tags:"定时任务" summary:"查询任务列表" dc:"分页查询定时任务列表，支持按任务名称、分组、状态筛选"`
	Name     string `json:"name" dc:"任务名称，支持模糊查询" eg:"会话清理"`
	Group    string `json:"group" dc:"任务分组" eg:"system"`
	Status   *int   `json:"status" dc:"任务状态：1=启用 0=禁用，不传则查询全部" eg:"1"`
	Page     int    `json:"page" v:"required|min:1" dc:"页码" eg:"1"`
	PageSize int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量" eg:"10"`
}

type JobListRes struct {
	Items []*JobItem `json:"items" dc:"任务列表"`
	Total int        `json:"total" dc:"总数"`
}

type JobItem struct {
	Id          uint64 `json:"id" dc:"任务ID"`
	Name        string `json:"name" dc:"任务名称"`
	Group       string `json:"group" dc:"任务分组"`
	Command     string `json:"command" dc:"执行指令"`
	CronExpr    string `json:"cronExpr" dc:"Cron表达式"`
	Description string `json:"description" dc:"任务描述"`
	Status      int    `json:"status" dc:"状态：1=启用 0=禁用"`
	Singleton   int    `json:"singleton" dc:"执行模式：1=单例 0=并行"`
	MaxTimes    int    `json:"maxTimes" dc:"最大执行次数"`
	ExecTimes   int    `json:"execTimes" dc:"已执行次数"`
	IsSystem    int    `json:"isSystem" dc:"是否系统任务：1=是 0=否"`
	CreateBy    string `json:"createBy" dc:"创建者"`
	CreateTime  string `json:"createTime" dc:"创建时间"`
	UpdateBy    string `json:"updateBy" dc:"更新者"`
	UpdateTime  string `json:"updateTime" dc:"更新时间"`
	Remark      string `json:"remark" dc:"备注"`
}
