package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// System Info API

type GetInfoReq struct {
	g.Meta `path:"/system/info" method:"get" tags:"系统信息" summary:"获取系统运行信息"`
}

type GetInfoRes struct {
	GoVersion   string `json:"goVersion" dc:"Go版本"`
	GfVersion   string `json:"gfVersion" dc:"GoFrame版本"`
	Os          string `json:"os" dc:"操作系统"`
	Arch        string `json:"arch" dc:"系统架构"`
	DbVersion   string `json:"dbVersion" dc:"数据库版本"`
	StartTime   string `json:"startTime" dc:"系统启动时间"`
	RunDuration string `json:"runDuration" dc:"系统运行时长"`
}
