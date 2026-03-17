package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Option Select API

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
