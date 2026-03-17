package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg ReadAll API

type ReadAllReq struct {
	g.Meta `path:"/user/message/read-all" method:"put" tags:"用户消息" summary:"标记全部消息已读"`
}

type ReadAllRes struct{}
