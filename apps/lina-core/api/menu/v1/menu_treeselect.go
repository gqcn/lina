package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Menu TreeSelect API - for role menu selection dropdown

type TreeSelectReq struct {
	g.Meta `path:"/menu/treeselect" method:"get" tags:"菜单管理" summary:"获取菜单下拉树" dc:"获取菜单下拉树，用于角色分配菜单时选择。过滤掉按钮类型的菜单"`
}

// MenuTreeNode represents a node in the tree select
type MenuTreeNode struct {
	Id       int             `json:"id" dc:"菜单ID" eg:"1"`
	ParentId int             `json:"parentId" dc:"父菜单ID" eg:"0"`
	Label    string          `json:"label" dc:"菜单名称" eg:"系统管理"`
	Children []*MenuTreeNode `json:"children" dc:"子菜单" eg:"[]"`
}

type TreeSelectRes struct {
	List []*MenuTreeNode `json:"list" dc:"菜单树形列表" eg:"[]"`
}