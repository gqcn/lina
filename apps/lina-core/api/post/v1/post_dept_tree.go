package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Dept Tree API

type DeptTreeReq struct {
	g.Meta `path:"/post/dept-tree" method:"get" tags:"岗位管理" summary:"获取岗位筛选部门树" dc:"获取部门树结构数据，用于岗位列表页面按部门筛选岗位，每个节点包含部门下的岗位数量"`
}

// DeptTreeRes 部门树响应
type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树" eg:"[]"`
}

// DeptTreeNode 部门树节点
type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID" eg:"100"`
	Label     string          `json:"label" dc:"部门名称" eg:"技术部"`
	PostCount int             `json:"postCount" dc:"该部门下的岗位数量" eg:"5"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表" eg:"[]"`
}
