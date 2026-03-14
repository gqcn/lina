// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDictData is the golang structure for table sys_dict_data.
type SysDictData struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	DictType  string      `json:"dictType"  orm:"dict_type"  description:""`
	Label     string      `json:"label"     orm:"label"      description:""`
	Value     string      `json:"value"     orm:"value"      description:""`
	Sort      int         `json:"sort"      orm:"sort"       description:""`
	TagStyle  string      `json:"tagStyle"  orm:"tag_style"  description:""`
	CssClass  string      `json:"cssClass"  orm:"css_class"  description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Remark    string      `json:"remark"    orm:"remark"     description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
}
