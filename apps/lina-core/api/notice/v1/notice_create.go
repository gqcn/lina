package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Create API

type CreateReq struct {
	g.Meta  `path:"/notice" method:"post" tags:"通知公告" summary:"创建通知公告"`
	Title   string `json:"title" v:"required#请输入公告标题" dc:"公告标题"`
	Type    int    `json:"type" v:"required|in:1,2#请选择公告类型|公告类型不正确" dc:"公告类型：1=通知 2=公告"`
	Content string `json:"content" v:"required#请输入公告内容" dc:"公告内容"`
	Status  *int   `json:"status" d:"0" dc:"公告状态：0=草稿 1=已发布"`
	Remark  string `json:"remark" dc:"备注"`
}

type CreateRes struct {
	Id int64 `json:"id" dc:"公告ID"`
}
