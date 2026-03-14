// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPost is the golang structure of table sys_post for DAO operations like Where/Data.
type SysPost struct {
	g.Meta    `orm:"table:sys_post, do:true"`
	Id        any         //
	DeptId    any         //
	Code      any         //
	Name      any         //
	Sort      any         //
	Status    any         //
	Remark    any         //
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
	DeletedAt *gtime.Time //
}
