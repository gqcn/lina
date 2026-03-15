package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta `path:"/dept" method:"get" tags:"部门管理" summary:"获取部门列表"`
	Name   string `json:"name" dc:"按部门名称筛选"`
	Status *int   `json:"status" dc:"按状态筛选"`
}

type ListRes struct {
	List []*entity.SysDept `json:"list" dc:"部门列表"`
}

type CreateReq struct {
	g.Meta   `path:"/dept" method:"post" tags:"部门管理" summary:"创建部门"`
	ParentId int    `json:"parentId" d:"0" dc:"父级部门ID"`
	Name     string `json:"name" v:"required#请输入部门名称" dc:"部门名称"`
	Code     string `json:"code" dc:"部门编码（唯一）"`
	OrderNum *int   `json:"orderNum" d:"0" dc:"排序号"`
	Leader   *int   `json:"leader" dc:"负责人用户ID"`
	Phone    string `json:"phone" dc:"联系电话"`
	Email    string `json:"email" dc:"邮箱"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark   string `json:"remark" dc:"备注"`
}

type CreateRes struct {
	Id int `json:"id" dc:"部门ID"`
}

type GetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"部门管理" summary:"获取部门详情"`
	Id     int `json:"id" v:"required" dc:"部门ID"`
}

type GetRes struct {
	*entity.SysDept `dc:"部门信息"`
}

type UpdateReq struct {
	g.Meta   `path:"/dept/{id}" method:"put" tags:"部门管理" summary:"更新部门"`
	Id       int     `json:"id" v:"required" dc:"部门ID"`
	ParentId *int    `json:"parentId" dc:"父级部门ID"`
	Name     *string `json:"name" dc:"部门名称"`
	Code     *string `json:"code" dc:"部门编码（唯一）"`
	OrderNum *int    `json:"orderNum" dc:"排序号"`
	Leader   *int    `json:"leader" dc:"负责人用户ID"`
	Phone    *string `json:"phone" dc:"联系电话"`
	Email    *string `json:"email" dc:"邮箱"`
	Status   *int    `json:"status" dc:"状态"`
	Remark   *string `json:"remark" dc:"备注"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/dept/{id}" method:"delete" tags:"部门管理" summary:"删除部门"`
	Id     int `json:"id" v:"required" dc:"部门ID"`
}

type DeleteRes struct{}

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

// ExcludeReq returns dept list excluding a node and its children.
type ExcludeReq struct {
	g.Meta `path:"/dept/exclude/{id}" method:"get" tags:"部门管理" summary:"获取排除节点后的部门列表"`
	Id     int `json:"id" v:"required" dc:"需排除的部门ID"`
}

type ExcludeRes struct {
	List []*entity.SysDept `json:"list" dc:"部门列表"`
}

// UsersReq returns users belonging to a dept (for leader selection).
// When Id=0, returns all users. When Id>0, returns users in the dept and all its sub-depts.
type UsersReq struct {
	g.Meta  `path:"/dept/{id}/users" method:"get" tags:"部门管理" summary:"获取部门用户列表"`
	Id      int    `json:"id" dc:"部门ID，0表示所有用户"`
	Keyword string `json:"keyword" dc:"按用户名或昵称搜索"`
	Limit   int    `json:"limit" d:"10" dc:"最大返回条数"`
}

type DeptUser struct {
	Id       int    `json:"id" dc:"用户ID"`
	Username string `json:"username" dc:"用户名"`
	Nickname string `json:"nickname" dc:"昵称"`
}

type UsersRes struct {
	List []*DeptUser `json:"list" dc:"用户列表"`
}
