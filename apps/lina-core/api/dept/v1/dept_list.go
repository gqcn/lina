package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dept List API

type ListReq struct {
	g.Meta `path:"/dept" method:"get" tags:"部门管理" summary:"获取部门列表"`
	Name   string `json:"name" dc:"按部门名称筛选"`
	Status *int   `json:"status" dc:"按状态筛选"`
}

type ListRes struct {
	List []*entity.SysDept `json:"list" dc:"部门列表"`
}
