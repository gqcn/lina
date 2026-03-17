package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Notice List API

type ListReq struct {
	g.Meta   `path:"/notice" method:"get" tags:"通知公告" summary:"获取通知公告列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Title    string `json:"title" dc:"按标题筛选"`
	Type     int    `json:"type" dc:"按类型筛选"`
	CreatedBy string `json:"createdBy" dc:"按创建人用户名筛选"`
}

type ListRes struct {
	List  []*ListItem `json:"list" dc:"通知公告列表"`
	Total int         `json:"total" dc:"总条数"`
}

type ListItem struct {
	*entity.SysNotice
	CreatedByName string `json:"createdByName" dc:"创建人名称"`
}
