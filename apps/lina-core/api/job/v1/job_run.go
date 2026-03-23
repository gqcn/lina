package v1

import "github.com/gogf/gf/v2/frame/g"

type JobRunReq struct {
	g.Meta `path:"/job/run" method:"post" tags:"定时任务" summary:"手动执行任务" dc:"立即执行一次定时任务"`
	Id     uint64 `json:"id" v:"required|min:1" dc:"任务ID" eg:"1"`
}

type JobRunRes struct{}
