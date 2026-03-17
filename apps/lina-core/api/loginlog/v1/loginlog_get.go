package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Get API

type GetReq struct {
	g.Meta `path:"/loginlog/{id}" method:"get" tags:"登录日志" summary:"获取登录日志详情"`
	Id     int `json:"id" v:"required" dc:"登录日志ID"`
}

type GetRes struct {
	*entity.SysLoginLog
}
