package v1

import "github.com/gogf/gf/v2/frame/g"

type JobCreateReq struct {
	g.Meta      `path:"/job/create" method:"post" tags:"定时任务" summary:"创建任务" dc:"创建新的定时任务"`
	Name        string `json:"name" v:"required|length:1,64" dc:"任务名称" eg:"数据备份"`
	Group       string `json:"group" v:"required|length:1,64" dc:"任务分组" eg:"default"`
	Command     string `json:"command" v:"required|length:1,500" dc:"执行指令" eg:"echo 'backup'"`
	CronExpr    string `json:"cronExpr" v:"required|length:1,255" dc:"Cron表达式" eg:"0 0 2 * * *"`
	Description string `json:"description" v:"length:0,500" dc:"任务描述" eg:"每天凌晨2点执行数据备份"`
	Status      int    `json:"status" v:"required|in:0,1" dc:"状态：1=启用 0=禁用" eg:"1"`
	Singleton   int    `json:"singleton" v:"required|in:0,1" dc:"执行模式：1=单例 0=并行" eg:"1"`
	MaxTimes    int    `json:"maxTimes" v:"min:0" dc:"最大执行次数，0表示无限制" eg:"0"`
	Remark      string `json:"remark" v:"length:0,500" dc:"备注" eg:""`
}

type JobCreateRes struct{}
