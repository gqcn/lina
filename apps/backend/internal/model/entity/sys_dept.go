// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDept is the golang structure for table sys_dept.
type SysDept struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	ParentId  int         `json:"parentId"  orm:"parent_id"  description:""`
	Ancestors string      `json:"ancestors" orm:"ancestors"  description:""`
	Name      string      `json:"name"      orm:"name"       description:""`
	OrderNum  int         `json:"orderNum"  orm:"order_num"  description:""`
	Leader    int         `json:"leader"    orm:"leader"     description:""`
	Phone     string      `json:"phone"     orm:"phone"      description:""`
	Email     string      `json:"email"     orm:"email"      description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Remark    string      `json:"remark"    orm:"remark"     description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
	Code      string      `json:"code"      orm:"code"       description:""`
}
