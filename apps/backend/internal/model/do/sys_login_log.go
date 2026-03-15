// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLoginLog is the golang structure of table sys_login_log for DAO operations like Where/Data.
type SysLoginLog struct {
	g.Meta    `orm:"table:sys_login_log, do:true"`
	Id        any         //
	UserName  any         //
	Status    any         //
	Ip        any         //
	Browser   any         //
	Os        any         //
	Msg       any         //
	LoginTime *gtime.Time //
}
