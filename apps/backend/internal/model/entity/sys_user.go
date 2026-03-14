// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUser is the golang structure for table sys_user.
type SysUser struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	Username  string      `json:"username"  orm:"username"   description:""`
	Password  string      `json:"password"  orm:"password"   description:""`
	Nickname  string      `json:"nickname"  orm:"nickname"   description:""`
	Email     string      `json:"email"     orm:"email"      description:""`
	Phone     string      `json:"phone"     orm:"phone"      description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Remark    string      `json:"remark"    orm:"remark"     description:""`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
	DeletedAt *gtime.Time `json:"deletedAt" orm:"deleted_at" description:""`
	Avatar    string      `json:"avatar"    orm:"avatar"     description:""`
	LoginDate *gtime.Time `json:"loginDate" orm:"login_date" description:""`
	Sex       int         `json:"sex"       orm:"sex"        description:""`
}
