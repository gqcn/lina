package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// UserMsg List API

type ListReq struct {
	g.Meta   `path:"/user/message" method:"get" tags:"用户消息" summary:"获取消息列表" dc:"分页查询当前登录用户的消息列表，包括已读和未读消息"`
	PageNum  int `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize int `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
}

type ListRes struct {
	List  []*entity.SysUserMessage `json:"list" dc:"消息列表"`
	Total int                      `json:"total" dc:"总条数" eg:"20"`
}
