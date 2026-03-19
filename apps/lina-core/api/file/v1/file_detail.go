package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// File Detail API

type DetailReq struct {
	g.Meta `path:"/file/detail/{id}" method:"get" tags:"文件管理" summary:"获取文件详情" dc:"根据文件ID查询文件完整详细信息，包括文件基本信息、上传者名称以及文件的使用场景列表"`
	Id     int64 `json:"id" v:"required" dc:"文件ID" eg:"1"`
}

type DetailRes struct {
	*entity.SysFile
	CreatedByName string             `json:"createdByName" dc:"上传者用户名" eg:"admin"`
	UsageScenes   []*DetailUsageItem `json:"usageScenes" dc:"文件使用场景列表"`
}

type DetailUsageItem struct {
	Scene     string `json:"scene" dc:"使用场景标识：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他" eg:"avatar"`
	Label     string `json:"label" dc:"使用场景名称" eg:"用户头像"`
	CreatedAt string `json:"createdAt" dc:"关联时间" eg:"2026-01-01 12:00:00"`
}
