package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta   `path:"/post" method:"get" tags:"Post" summary:"Get post list"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	DeptId   *int   `json:"deptId" dc:"Filter by dept ID"`
	Code     string `json:"code" dc:"Filter by post code"`
	Name     string `json:"name" dc:"Filter by post name"`
	Status   *int   `json:"status" dc:"Filter by status"`
}

type ListRes struct {
	List  []*entity.SysPost `json:"list" dc:"Post list"`
	Total int               `json:"total" dc:"Total count"`
}

type CreateReq struct {
	g.Meta `path:"/post" method:"post" tags:"Post" summary:"Create post"`
	DeptId int    `json:"deptId" v:"required#请选择所属部门" dc:"Dept ID"`
	Code   string `json:"code" v:"required#请输入岗位编码" dc:"Post code"`
	Name   string `json:"name" v:"required#请输入岗位名称" dc:"Post name"`
	Sort   *int   `json:"sort" d:"0" dc:"Sort order"`
	Status *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled"`
	Remark string `json:"remark" dc:"Remark"`
}

type CreateRes struct {
	Id int `json:"id" dc:"Post ID"`
}

type GetReq struct {
	g.Meta `path:"/post/{id}" method:"get" tags:"Post" summary:"Get post detail"`
	Id     int `json:"id" v:"required" dc:"Post ID"`
}

type GetRes struct {
	*entity.SysPost `dc:"Post info"`
}

type UpdateReq struct {
	g.Meta `path:"/post/{id}" method:"put" tags:"Post" summary:"Update post"`
	Id     int     `json:"id" v:"required" dc:"Post ID"`
	DeptId *int    `json:"deptId" dc:"Dept ID"`
	Code   *string `json:"code" dc:"Post code"`
	Name   *string `json:"name" dc:"Post name"`
	Sort   *int    `json:"sort" dc:"Sort order"`
	Status *int    `json:"status" dc:"Status"`
	Remark *string `json:"remark" dc:"Remark"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/post/{ids}" method:"delete" tags:"Post" summary:"Delete post(s)"`
	Ids    string `json:"ids" v:"required" dc:"Post ID(s), comma separated"`
}

type DeleteRes struct{}

type ExportReq struct {
	g.Meta `path:"/post/export" method:"get" tags:"Post" summary:"Export posts to Excel" operLog:"4"`
	DeptId *int   `json:"deptId" dc:"Filter by dept ID"`
	Code   string `json:"code" dc:"Filter by post code"`
	Name   string `json:"name" dc:"Filter by post name"`
	Status *int   `json:"status" dc:"Filter by status"`
}

type ExportRes struct{}

type DeptTreeReq struct {
	g.Meta `path:"/post/dept-tree" method:"get" tags:"Post" summary:"Get dept tree for post filter"`
}

type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"Dept tree"`
}

type DeptTreeNode struct {
	Id        int             `json:"id" dc:"Dept ID"`
	Label     string          `json:"label" dc:"Dept name"`
	PostCount int             `json:"postCount" dc:"Post count"`
	Children  []*DeptTreeNode `json:"children" dc:"Child depts"`
}

type OptionSelectReq struct {
	g.Meta `path:"/post/option-select" method:"get" tags:"Post" summary:"Get post options by dept"`
	DeptId *int `json:"deptId" dc:"Dept ID"`
}

type PostOption struct {
	PostId   int    `json:"postId" dc:"Post ID"`
	PostName string `json:"postName" dc:"Post name"`
}

type OptionSelectRes struct {
	List []*PostOption `json:"list" dc:"Post options"`
}
