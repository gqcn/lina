package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dept Get API

type GetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"部门管理" summary:"获取部门详情"`
	Id     int `json:"id" v:"required" dc:"部门ID"`
}

type GetRes struct {
	*entity.SysDept `dc:"部门信息"`
}
