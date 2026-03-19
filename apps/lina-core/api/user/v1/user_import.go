package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Import API

type ImportReq struct {
	g.Meta `path:"/user/import" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"导入用户数据" dc:"通过Excel文件批量导入用户数据，需使用系统提供的导入模板"`
}

type ImportRes struct {
	Success  int              `json:"success" dc:"成功条数" eg:"10"`
	Fail     int              `json:"fail" dc:"失败条数" eg:"2"`
	FailList []ImportFailItem `json:"failList" dc:"失败详情"`
}

type ImportFailItem struct {
	Row    int    `json:"row" dc:"行号" eg:"3"`
	Reason string `json:"reason" dc:"失败原因" eg:"用户名已存在"`
}

type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"用户管理" summary:"下载导入模板" dc:"下载用户导入Excel模板文件，包含必填字段和数据格式说明"`
}

type ImportTemplateRes struct{}
