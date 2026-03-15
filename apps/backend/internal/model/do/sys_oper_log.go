// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysOperLog is the golang structure of table sys_oper_log for DAO operations like Where/Data.
type SysOperLog struct {
	g.Meta        `orm:"table:sys_oper_log, do:true"`
	Id            any         //
	Title         any         //
	OperSummary   any         //
	OperType      any         //
	Method        any         //
	RequestMethod any         //
	OperName      any         //
	OperUrl       any         //
	OperIp        any         //
	OperParam     any         //
	JsonResult    any         //
	Status        any         //
	ErrorMsg      any         //
	CostTime      any         //
	OperTime      *gtime.Time //
}
