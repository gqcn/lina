package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// File Usage Scenes API

type UsageScenesReq struct {
	g.Meta `path:"/file/scenes" method:"get" tags:"文件管理" summary:"获取文件使用场景列表" dc:"查询所有已使用的文件使用场景标识列表，用于文件列表页面的使用场景筛选下拉框"`
}

type UsageScenesRes struct {
	List []*UsageSceneItem `json:"list" dc:"使用场景列表"`
}

type UsageSceneItem struct {
	Value string `json:"value" dc:"使用场景标识" eg:"avatar"`
	Label string `json:"label" dc:"使用场景名称" eg:"用户头像"`
}
