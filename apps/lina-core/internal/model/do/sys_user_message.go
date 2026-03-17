// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUserMessage is the golang structure of table sys_user_message for DAO operations like Where/Data.
type SysUserMessage struct {
	g.Meta     `orm:"table:sys_user_message, do:true"`
	Id         any         // 消息ID
	UserId     any         // 接收用户ID
	Title      any         // 消息标题
	Type       any         // 消息类型（1通知 2公告）
	SourceType any         // 来源类型
	SourceId   any         // 来源ID
	IsRead     any         // 是否已读（0未读 1已读）
	ReadAt     *gtime.Time // 阅读时间
	CreatedAt  *gtime.Time // 创建时间
}
