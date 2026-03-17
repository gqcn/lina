package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Update API

type UpdateReq struct {
	g.Meta  `path:"/notice/{id}" method:"put" tags:"通知公告" summary:"更新通知公告"`
	Id      int64   `json:"id" v:"required" dc:"公告ID"`
	Title   *string `json:"title" dc:"公告标题"`
	Type    *int    `json:"type" dc:"公告类型"`
	Content *string `json:"content" dc:"公告内容"`
	Status  *int    `json:"status" dc:"公告状态"`
	Remark  *string `json:"remark" dc:"备注"`
}

type UpdateRes struct{}
