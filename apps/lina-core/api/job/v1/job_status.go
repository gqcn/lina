package v1

import "github.com/gogf/gf/v2/frame/g"

type JobStatusReq struct {
	g.Meta `path:"/job/status" method:"put" tags:"定时任务" summary:"更新任务状态" dc:"启用或禁用定时任务"`
	Id     uint64 `json:"id" v:"required|min:1" dc:"任务ID" eg:"1"`
	Status int    `json:"status" v:"required|in:0,1" dc:"状态：1=启用 0=禁用" eg:"1"`
}

type JobStatusRes struct{}
