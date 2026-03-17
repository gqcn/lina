package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg Count API

type CountReq struct {
	g.Meta `path:"/user/message/count" method:"get" tags:"用户消息" summary:"获取未读消息数量"`
}

type CountRes struct {
	Count int `json:"count" dc:"未读消息数量"`
}
