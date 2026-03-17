package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dept Exclude API

// ExcludeReq returns dept list excluding a node and its children.
type ExcludeReq struct {
	g.Meta `path:"/dept/exclude/{id}" method:"get" tags:"部门管理" summary:"获取排除节点后的部门列表"`
	Id     int `json:"id" v:"required" dc:"需排除的部门ID"`
}

type ExcludeRes struct {
	List []*entity.SysDept `json:"list" dc:"部门列表"`
}
