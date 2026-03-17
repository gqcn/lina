package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Import API

type ImportReq struct {
	g.Meta `path:"/user/import" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"导入用户数据"`
}

type ImportRes struct {
	Success  int              `json:"success" dc:"成功条数"`
	Fail     int              `json:"fail" dc:"失败条数"`
	FailList []ImportFailItem `json:"failList" dc:"失败详情"`
}

type ImportFailItem struct {
	Row    int    `json:"row" dc:"行号"`
	Reason string `json:"reason" dc:"失败原因"`
}

type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"用户管理" summary:"下载导入模板"`
}

type ImportTemplateRes struct{}
