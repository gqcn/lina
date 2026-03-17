package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Tree API

// TreeReq returns dept tree for TreeSelect component.
type TreeReq struct {
	g.Meta `path:"/dept/tree" method:"get" tags:"部门管理" summary:"获取部门树"`
}

type TreeNode struct {
	Id       int         `json:"id" dc:"部门ID"`
	Label    string      `json:"label" dc:"部门名称"`
	Children []*TreeNode `json:"children" dc:"子部门列表"`
}

type TreeRes struct {
	List []*TreeNode `json:"list" dc:"部门树"`
}
