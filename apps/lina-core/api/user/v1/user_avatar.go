package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Avatar API

type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"上传并更新头像"`
}

type UpdateAvatarRes struct {
	Url string `json:"url" dc:"头像地址"`
}
