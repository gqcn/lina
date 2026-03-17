package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta   `path:"/post" method:"get" tags:"岗位管理" summary:"获取岗位列表"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	DeptId   *int   `json:"deptId" dc:"按部门ID筛选"`
	Code     string `json:"code" dc:"按岗位编码筛选"`
	Name     string `json:"name" dc:"按岗位名称筛选"`
	Status   *int   `json:"status" dc:"按状态筛选"`
}

type ListRes struct {
	List  []*entity.SysPost `json:"list" dc:"岗位列表"`
	Total int               `json:"total" dc:"总条数"`
}

type CreateReq struct {
	g.Meta `path:"/post" method:"post" tags:"岗位管理" summary:"创建岗位"`
	DeptId int    `json:"deptId" v:"required#请选择所属部门" dc:"部门ID"`
	Code   string `json:"code" v:"required#请输入岗位编码" dc:"岗位编码"`
	Name   string `json:"name" v:"required#请输入岗位名称" dc:"岗位名称"`
	Sort   *int   `json:"sort" d:"0" dc:"排序号"`
	Status *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark string `json:"remark" dc:"备注"`
}

type CreateRes struct {
	Id int `json:"id" dc:"岗位ID"`
}

type GetReq struct {
	g.Meta `path:"/post/{id}" method:"get" tags:"岗位管理" summary:"获取岗位详情"`
	Id     int `json:"id" v:"required" dc:"岗位ID"`
}

type GetRes struct {
	*entity.SysPost `dc:"岗位信息"`
}

type UpdateReq struct {
	g.Meta `path:"/post/{id}" method:"put" tags:"岗位管理" summary:"更新岗位"`
	Id     int     `json:"id" v:"required" dc:"岗位ID"`
	DeptId *int    `json:"deptId" dc:"部门ID"`
	Code   *string `json:"code" dc:"岗位编码"`
	Name   *string `json:"name" dc:"岗位名称"`
	Sort   *int    `json:"sort" dc:"排序号"`
	Status *int    `json:"status" dc:"状态"`
	Remark *string `json:"remark" dc:"备注"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/post/{ids}" method:"delete" tags:"岗位管理" summary:"删除岗位"`
	Ids    string `json:"ids" v:"required" dc:"岗位ID，多个用逗号分隔"`
}

type DeleteRes struct{}

type ExportReq struct {
	g.Meta `path:"/post/export" method:"get" tags:"岗位管理" summary:"导出岗位数据" operLog:"4"`
	DeptId *int   `json:"deptId" dc:"按部门ID筛选"`
	Code   string `json:"code" dc:"按岗位编码筛选"`
	Name   string `json:"name" dc:"按岗位名称筛选"`
	Status *int   `json:"status" dc:"按状态筛选"`
}

type ExportRes struct{}

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

type OptionSelectReq struct {
	g.Meta `path:"/post/option-select" method:"get" tags:"岗位管理" summary:"获取部门下岗位选项"`
	DeptId *int `json:"deptId" dc:"部门ID"`
}

type PostOption struct {
	PostId   int    `json:"postId" dc:"岗位ID"`
	PostName string `json:"postName" dc:"岗位名称"`
}

type OptionSelectRes struct {
	List []*PostOption `json:"list" dc:"岗位选项列表"`
}
