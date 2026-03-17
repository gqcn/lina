// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysLoginLog is the golang structure for table sys_login_log.
type SysLoginLog struct {
	Id        int         `json:"id"        orm:"id"         description:""`
	UserName  string      `json:"userName"  orm:"user_name"  description:""`
	Status    int         `json:"status"    orm:"status"     description:""`
	Ip        string      `json:"ip"        orm:"ip"         description:""`
	Browser   string      `json:"browser"   orm:"browser"    description:""`
	Os        string      `json:"os"        orm:"os"         description:""`
	Msg       string      `json:"msg"       orm:"msg"        description:""`
	LoginTime *gtime.Time `json:"loginTime" orm:"login_time" description:""`
}
