package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User DeptTree API

type DeptTreeReq struct {
	g.Meta `path:"/user/dept-tree" method:"get" tags:"用户管理" summary:"获取用户筛选部门树"`
}

type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID"`
	Label     string          `json:"label" dc:"部门名称"`
	UserCount int             `json:"userCount" dc:"部门用户数"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表"`
}

type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树"`
}
