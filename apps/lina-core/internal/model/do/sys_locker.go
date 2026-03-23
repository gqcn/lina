// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLocker is the golang structure of table sys_locker for DAO operations like Where/Data.
type SysLocker struct {
	g.Meta     `orm:"table:sys_locker, do:true"`
	Id         any         // 锁ID
	Name       any         // 锁名称
	Reason     any         // 锁定原因
	CreateTime *gtime.Time // 创建时间
	ExpireTime *gtime.Time // 过期时间
}
