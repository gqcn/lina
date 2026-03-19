// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysFileUsage is the golang structure for table sys_file_usage.
type SysFileUsage struct {
	Id        int64       `json:"id"        orm:"id"         description:"记录ID"`
	FileId    int64       `json:"fileId"    orm:"file_id"    description:"文件ID，关联 sys_file.id"`
	Scene     string      `json:"scene"     orm:"scene"      description:"使用场景标识：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他"`
	BizId     int64       `json:"bizId"     orm:"biz_id"     description:"关联的业务记录ID（如用户ID、通知ID等）"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`
}
