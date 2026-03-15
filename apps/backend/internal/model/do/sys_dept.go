// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDept is the golang structure of table sys_dept for DAO operations like Where/Data.
type SysDept struct {
	g.Meta    `orm:"table:sys_dept, do:true"`
	Id        any         //
	ParentId  any         //
	Ancestors any         //
	Name      any         //
	OrderNum  any         //
	Leader    any         //
	Phone     any         //
	Email     any         //
	Status    any         //
	Remark    any         //
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
	DeletedAt *gtime.Time //
	Code      any         //
}
