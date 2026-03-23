package v1

import "github.com/gogf/gf/v2/frame/g"

type JobDeleteReq struct {
	g.Meta `path:"/job/delete" method:"delete" tags:"定时任务" summary:"删除任务" dc:"删除定时任务，系统任务不可删除"`
	Ids    []uint64 `json:"ids" v:"required|length:1,100" dc:"任务ID列表" eg:"[1,2,3]"`
}

type JobDeleteRes struct{}
