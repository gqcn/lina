// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLocker is the golang structure for table sys_locker.
type SysLocker struct {
	Id         uint64      `json:"id"         orm:"id"          description:"锁ID"`
	Name       string      `json:"name"       orm:"name"        description:"锁名称"`
	Reason     string      `json:"reason"     orm:"reason"      description:"锁定原因"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
	ExpireTime *gtime.Time `json:"expireTime" orm:"expire_time" description:"过期时间"`
}
