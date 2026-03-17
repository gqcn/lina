package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// User List API

type ListReq struct {
	g.Meta         `path:"/user" method:"get" tags:"用户管理" summary:"获取用户列表"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Username       string `json:"username" dc:"按用户名筛选"`
	Nickname       string `json:"nickname" dc:"按昵称筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	Phone          string `json:"phone" dc:"按手机号筛选"`
	Sex            *int   `json:"sex" dc:"按性别筛选"`
	DeptId         *int   `json:"deptId" dc:"按部门ID筛选"`
	BeginTime      string `json:"beginTime" dc:"按创建时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按创建时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,username,nickname,phone,email,status,created_at"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc"`
}

type ListItem struct {
	*entity.SysUser
	DeptId   int    `json:"deptId" dc:"部门ID"`
	DeptName string `json:"deptName" dc:"部门名称"`
}

type ListRes struct {
	List  []*ListItem `json:"list" dc:"用户列表"`
	Total int         `json:"total" dc:"总条数"`
}
