package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Clear API

type ClearReq struct {
	g.Meta `path:"/user/message/clear" method:"delete" tags:"用户消息" summary:"清空全部消息"`
}

type ClearRes struct{}
