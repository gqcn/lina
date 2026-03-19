package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Post Get API

type GetReq struct {
	g.Meta `path:"/post/{id}" method:"get" tags:"岗位管理" summary:"获取岗位详情" dc:"根据岗位ID获取岗位的详细信息"`
	Id     int `json:"id" v:"required" dc:"岗位ID" eg:"1"`
}

type GetRes struct {
	*entity.SysPost `dc:"岗位信息"`
}
