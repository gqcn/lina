package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta `path:"/dept" method:"get" tags:"Dept" summary:"Get dept list"`
	Name   string `json:"name" dc:"Filter by dept name"`
	Status *int   `json:"status" dc:"Filter by status"`
}

type ListRes struct {
	List []*entity.SysDept `json:"list" dc:"Dept list"`
}

type CreateReq struct {
	g.Meta   `path:"/dept" method:"post" tags:"Dept" summary:"Create dept"`
	ParentId int    `json:"parentId" d:"0" dc:"Parent dept ID"`
	Name     string `json:"name" v:"required#请输入部门名称" dc:"Dept name"`
	OrderNum *int   `json:"orderNum" d:"0" dc:"Sort order"`
	Leader   *int   `json:"leader" dc:"Leader user ID"`
	Phone    string `json:"phone" dc:"Phone"`
	Email    string `json:"email" dc:"Email"`
	Status   *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled"`
	Remark   string `json:"remark" dc:"Remark"`
}

type CreateRes struct {
	Id int `json:"id" dc:"Dept ID"`
}

type GetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"Dept" summary:"Get dept detail"`
	Id     int `json:"id" v:"required" dc:"Dept ID"`
}

type GetRes struct {
	*entity.SysDept `dc:"Dept info"`
}

type UpdateReq struct {
	g.Meta   `path:"/dept/{id}" method:"put" tags:"Dept" summary:"Update dept"`
	Id       int     `json:"id" v:"required" dc:"Dept ID"`
	ParentId *int    `json:"parentId" dc:"Parent dept ID"`
	Name     *string `json:"name" dc:"Dept name"`
	OrderNum *int    `json:"orderNum" dc:"Sort order"`
	Leader   *int    `json:"leader" dc:"Leader user ID"`
	Phone    *string `json:"phone" dc:"Phone"`
	Email    *string `json:"email" dc:"Email"`
	Status   *int    `json:"status" dc:"Status"`
	Remark   *string `json:"remark" dc:"Remark"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/dept/{id}" method:"delete" tags:"Dept" summary:"Delete dept"`
	Id     int `json:"id" v:"required" dc:"Dept ID"`
}

type DeleteRes struct{}

// TreeReq returns dept tree for TreeSelect component.
type TreeReq struct {
	g.Meta `path:"/dept/tree" method:"get" tags:"Dept" summary:"Get dept tree"`
}

type TreeNode struct {
	Id       int         `json:"id" dc:"Dept ID"`
	Label    string      `json:"label" dc:"Dept name"`
	Children []*TreeNode `json:"children" dc:"Child depts"`
}

type TreeRes struct {
	List []*TreeNode `json:"list" dc:"Dept tree"`
}

// ExcludeReq returns dept list excluding a node and its children.
type ExcludeReq struct {
	g.Meta `path:"/dept/exclude/{id}" method:"get" tags:"Dept" summary:"Get dept list excluding node"`
	Id     int `json:"id" v:"required" dc:"Dept ID to exclude"`
}

type ExcludeRes struct {
	List []*entity.SysDept `json:"list" dc:"Dept list"`
}

// UsersReq returns users belonging to a dept (for leader selection).
type UsersReq struct {
	g.Meta `path:"/dept/{id}/users" method:"get" tags:"Dept" summary:"Get dept users"`
	Id     int `json:"id" v:"required" dc:"Dept ID"`
}

type DeptUser struct {
	Id       int    `json:"id" dc:"User ID"`
	Username string `json:"username" dc:"Username"`
	Nickname string `json:"nickname" dc:"Nickname"`
}

type UsersRes struct {
	List []*DeptUser `json:"list" dc:"User list"`
}
