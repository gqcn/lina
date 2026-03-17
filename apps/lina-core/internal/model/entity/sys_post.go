// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPost is the golang structure for table sys_post.
type SysPost struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	DeptId    int         `json:"deptId"    orm:"dept_id"    description:""`
	Code      string      `json:"code"      orm:"code"       description:""`
	Name      string      `json:"name"      orm:"name"       description:""`
	Sort      int         `json:"sort"      orm:"sort"       description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Remark    string      `json:"remark"    orm:"remark"     description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
