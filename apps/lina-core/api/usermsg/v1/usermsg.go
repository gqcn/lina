package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// User Message Count API

type CountReq struct {
	g.Meta `path:"/user/message/count" method:"get" tags:"用户消息" summary:"获取未读消息数量"`
}

type CountRes struct {
	Count int `json:"count" dc:"未读消息数量"`
}

// User Message List API

type ListReq struct {
	g.Meta   `path:"/user/message" method:"get" tags:"用户消息" summary:"获取消息列表"`
	PageNum  int `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
}

type ListRes struct {
	List  []*entity.SysUserMessage `json:"list" dc:"消息列表"`
	Total int                      `json:"total" dc:"总条数"`
}

// User Message Read API

type ReadReq struct {
	g.Meta `path:"/user/message/{id}/read" method:"put" tags:"用户消息" summary:"标记消息已读"`
	Id     int64 `json:"id" v:"required" dc:"消息ID"`
}

type ReadRes struct{}

// User Message Read All API

type ReadAllReq struct {
	g.Meta `path:"/user/message/read-all" method:"put" tags:"用户消息" summary:"标记全部消息已读"`
}

type ReadAllRes struct{}

// User Message Delete API

type DeleteReq struct {
	g.Meta `path:"/user/message/{id}" method:"delete" tags:"用户消息" summary:"删除单条消息"`
	Id     int64 `json:"id" v:"required" dc:"消息ID"`
}

type DeleteRes struct{}

// User Message Clear API

type ClearReq struct {
	g.Meta `path:"/user/message/clear" method:"delete" tags:"用户消息" summary:"清空全部消息"`
}

type ClearRes struct{}
