package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Dept Tree API

type DeptTreeReq struct {
	g.Meta `path:"/post/dept-tree" method:"get" tags:"岗位管理" summary:"获取岗位筛选部门树"`
}

type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树"`
}

type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID"`
	Label     string          `json:"label" dc:"部门名称"`
	PostCount int             `json:"postCount" dc:"岗位数量"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表"`
}
