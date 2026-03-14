// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictType is the golang structure for table sys_dict_type.
type SysDictType struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	Name      string      `json:"name"      orm:"name"       description:""`
	Type      string      `json:"type"      orm:"type"       description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Remark    string      `json:"remark"    orm:"remark"     description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
