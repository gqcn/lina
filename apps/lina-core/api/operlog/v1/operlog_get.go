package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Get API

type GetReq struct {
	g.Meta `path:"/operlog/{id}" method:"get" tags:"操作日志" summary:"获取操作日志详情"`
	Id     int `json:"id" v:"required" dc:"操作日志ID"`
}

type GetRes struct {
	*entity.SysOperLog
}
