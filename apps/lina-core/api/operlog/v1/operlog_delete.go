package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Delete API

type DeleteReq struct {
	g.Meta `path:"/operlog/{ids}" method:"delete" tags:"操作日志" summary:"删除操作日志"`
	Ids    string `json:"ids" v:"required" dc:"日志ID，多个用逗号分隔"`
}

type DeleteRes struct {
	Deleted int `json:"deleted" dc:"删除记录数"`
}
