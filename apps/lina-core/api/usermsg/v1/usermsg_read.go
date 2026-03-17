package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Read API

type ReadReq struct {
	g.Meta `path:"/user/message/{id}/read" method:"put" tags:"用户消息" summary:"标记消息已读"`
	Id     int64 `json:"id" v:"required" dc:"消息ID"`
}

type ReadRes struct{}
