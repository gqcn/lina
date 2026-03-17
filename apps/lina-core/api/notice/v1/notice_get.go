package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Notice Get API

type GetReq struct {
	g.Meta `path:"/notice/{id}" method:"get" tags:"通知公告" summary:"获取通知公告详情"`
	Id     int64 `json:"id" v:"required" dc:"公告ID"`
}

type GetRes struct {
	*entity.SysNotice
	CreatedByName string `json:"createdByName" dc:"创建人名称"`
}
