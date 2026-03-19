// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysFileUsage is the golang structure of table sys_file_usage for DAO operations like Where/Data.
type SysFileUsage struct {
	g.Meta    `orm:"table:sys_file_usage, do:true"`
	Id        any         // 记录ID
	FileId    any         // 文件ID，关联 sys_file.id
	Scene     any         // 使用场景标识：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他
	BizId     any         // 关联的业务记录ID（如用户ID、通知ID等）
	CreatedAt *gtime.Time // 创建时间
}
