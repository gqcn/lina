package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Post List API

type ListReq struct {
	g.Meta   `path:"/post" method:"get" tags:"岗位管理" summary:"获取岗位列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	DeptId   *int   `json:"deptId" dc:"按部门ID筛选"`
	Code     string `json:"code" dc:"按岗位编码筛选"`
	Name     string `json:"name" dc:"按岗位名称筛选"`
	Status   *int   `json:"status" dc:"按状态筛选"`
}

type ListRes struct {
	List  []*entity.SysPost `json:"list" dc:"岗位列表"`
	Total int               `json:"total" dc:"总条数"`
}
