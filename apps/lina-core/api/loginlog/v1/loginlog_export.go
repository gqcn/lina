package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Export API

type ExportReq struct {
	g.Meta         `path:"/loginlog/export" method:"get" tags:"登录日志" summary:"导出登录日志" operLog:"4"`
	UserName       string `json:"userName" dc:"按用户名筛选"`
	Ip             string `json:"ip" dc:"按IP地址筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	BeginTime      string `json:"beginTime" dc:"按登录时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按登录时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向"`
}

type ExportRes struct{}
