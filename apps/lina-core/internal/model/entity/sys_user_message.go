// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUserMessage is the golang structure for table sys_user_message.
type SysUserMessage struct {
	Id         int64       `json:"id"         orm:"id"          description:"消息ID"`
	UserId     int64       `json:"userId"     orm:"user_id"     description:"接收用户ID"`
	Title      string      `json:"title"      orm:"title"       description:"消息标题"`
	Type       int         `json:"type"       orm:"type"        description:"消息类型（1通知 2公告）"`
	SourceType string      `json:"sourceType" orm:"source_type" description:"来源类型"`
	SourceId   int64       `json:"sourceId"   orm:"source_id"   description:"来源ID"`
	IsRead     int         `json:"isRead"     orm:"is_read"     description:"是否已读（0未读 1已读）"`
	ReadAt     *gtime.Time `json:"readAt"     orm:"read_at"     description:"阅读时间"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
}
