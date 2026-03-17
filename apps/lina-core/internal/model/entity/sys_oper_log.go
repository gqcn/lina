// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysOperLog is the golang structure for table sys_oper_log.
type SysOperLog struct {
	Id            int         `json:"id"            orm:"id"             description:""`
	Title         string      `json:"title"         orm:"title"          description:""`
	OperSummary   string      `json:"operSummary"   orm:"oper_summary"   description:""`
	OperType      int         `json:"operType"      orm:"oper_type"      description:""`
	Method        string      `json:"method"        orm:"method"         description:""`
	RequestMethod string      `json:"requestMethod" orm:"request_method" description:""`
	OperName      string      `json:"operName"      orm:"oper_name"      description:""`
	OperUrl       string      `json:"operUrl"       orm:"oper_url"       description:""`
	OperIp        string      `json:"operIp"        orm:"oper_ip"        description:""`
	OperParam     string      `json:"operParam"     orm:"oper_param"     description:""`
	JsonResult    string      `json:"jsonResult"    orm:"json_result"    description:""`
	Status        int         `json:"status"        orm:"status"         description:""`
	ErrorMsg      string      `json:"errorMsg"      orm:"error_msg"      description:""`
	CostTime      int         `json:"costTime"      orm:"cost_time"      description:""`
	OperTime      *gtime.Time `json:"operTime"      orm:"oper_time"      description:""`
}
