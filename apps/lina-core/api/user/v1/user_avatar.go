package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User Avatar API

type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"上传并更新头像" dc:"上传头像图片文件并更新当前用户的头像，支持常见图片格式（jpg/png/gif等）"`
}

type UpdateAvatarRes struct {
	Url string `json:"url" dc:"头像地址" eg:"/upload/avatar/20250101120000.png"`
}
