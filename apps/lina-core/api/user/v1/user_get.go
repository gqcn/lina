package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// User Get API

type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"用户管理" summary:"获取用户详情"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
}

type GetRes struct {
	*entity.SysUser `dc:"用户信息"`
	DeptId          int    `json:"deptId" dc:"部门ID"`
	DeptName        string `json:"deptName" dc:"部门名称"`
	PostIds         []int  `json:"postIds" dc:"岗位ID列表"`
}
