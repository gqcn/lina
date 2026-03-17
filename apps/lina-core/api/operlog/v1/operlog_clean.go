package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Clean API

type CleanReq struct {
	g.Meta    `path:"/operlog/clean" method:"delete" tags:"操作日志" summary:"清空操作日志"`
	BeginTime string `json:"beginTime" dc:"清理起始时间"`
	EndTime   string `json:"endTime" dc:"清理截止时间"`
}

type CleanRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}
