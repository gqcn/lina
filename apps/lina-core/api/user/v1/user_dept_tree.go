package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeptTreeReq defines the request for querying the user department tree.
type DeptTreeReq struct {
	g.Meta `path:"/user/dept-tree" method:"get" tags:"用户管理" summary:"获取用户筛选部门树" dc:"获取部门树结构数据，用于用户列表页面按部门筛选用户，每个节点包含部门下的用户数量" permission:"system:user:query"`
}

// DeptTreeNode represents a node in the department tree for user filtering.
type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID" eg:"100"`
	Label     string          `json:"label" dc:"部门名称" eg:"技术部"`
	UserCount int             `json:"userCount" dc:"部门用户数" eg:"5"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表" eg:"[]"`
}

// DeptTreeRes is the response structure for department tree.
type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树" eg:"[]"`
}
