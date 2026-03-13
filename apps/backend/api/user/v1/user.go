package v1

import (
	"backend/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/user" method:"get" tags:"User" summary:"Get user list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Page size"`
	Username       string `json:"username" dc:"Filter by username"`
	Nickname       string `json:"nickname" dc:"Filter by nickname"`
	Status         *int   `json:"status" dc:"Filter by status"`
	Phone          string `json:"phone" dc:"Filter by phone"`
	BeginTime      string `json:"beginTime" dc:"Filter by created_at start time"`
	EndTime        string `json:"endTime" dc:"Filter by created_at end time"`
	OrderBy        string `json:"orderBy" dc:"Sort field: id,username,nickname,phone,email,status,created_at"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc"`
}

type ListRes struct {
	List  []*entity.SysUser `json:"list" dc:"User list"`
	Total int               `json:"total" dc:"Total count"`
}

type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"User" summary:"Create user"`
	Username string `json:"username" v:"required|length:2,64#请输入用户名|用户名长度为2-64个字符"`
	Password string `json:"password" v:"required|length:6,32#请输入密码|密码长度为6-32个字符"`
	Nickname string `json:"nickname" dc:"Nickname"`
	Email    string `json:"email" dc:"Email"`
	Phone    string `json:"phone" dc:"Phone"`
	Status   *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled"`
	Remark   string `json:"remark" dc:"Remark"`
}

type CreateRes struct {
	Id int `json:"id" dc:"User ID"`
}

type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"User" summary:"Get user detail"`
	Id     int `json:"id" v:"required" dc:"User ID"`
}

type GetRes struct {
	*entity.SysUser `dc:"User info"`
}

type UpdateReq struct {
	g.Meta   `path:"/user/{id}" method:"put" tags:"User" summary:"Update user"`
	Id       int     `json:"id" v:"required" dc:"User ID"`
	Username *string `json:"username" dc:"Username"`
	Password *string `json:"password" dc:"Password (empty means no change)"`
	Nickname *string `json:"nickname" dc:"Nickname"`
	Email    *string `json:"email" dc:"Email"`
	Phone    *string `json:"phone" dc:"Phone"`
	Status   *int    `json:"status" dc:"Status"`
	Remark   *string `json:"remark" dc:"Remark"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/user/{id}" method:"delete" tags:"User" summary:"Delete user"`
	Id     int `json:"id" v:"required" dc:"User ID"`
}

type DeleteRes struct{}

type UpdateStatusReq struct {
	g.Meta `path:"/user/{id}/status" method:"put" tags:"User" summary:"Update user status"`
	Id     int `json:"id" v:"required" dc:"User ID"`
	Status int `json:"status" v:"in:0,1#状态值无效" dc:"Status: 1=normal 0=disabled"`
}

type UpdateStatusRes struct{}

type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"User" summary:"Get current user profile"`
}

type GetProfileRes struct {
	*entity.SysUser `dc:"User profile"`
}

type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"User" summary:"Update current user profile"`
	Nickname *string `json:"nickname" dc:"Nickname"`
	Email    *string `json:"email" dc:"Email"`
	Phone    *string `json:"phone" dc:"Phone"`
	Password *string `json:"password" dc:"New password"`
}

type UpdateProfileRes struct{}

type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"User" summary:"Get current user info for frontend"`
}

type GetInfoRes struct {
	UserId   int      `json:"userId" dc:"User ID"`
	Username string   `json:"username" dc:"Username"`
	RealName string   `json:"realName" dc:"Real name"`
	Avatar   string   `json:"avatar" dc:"Avatar URL"`
	Roles    []string `json:"roles" dc:"User roles"`
	HomePath string   `json:"homePath" dc:"Home path"`
}

type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"User" summary:"Export users to Excel"`
	Ids    []int `json:"ids" dc:"Export specific user IDs"`
}

type ExportRes struct{}

type ImportReq struct {
	g.Meta `path:"/user/import" method:"post" mime:"multipart/form-data" tags:"User" summary:"Import users from Excel"`
}

type ImportRes struct {
	Success  int              `json:"success" dc:"Success count"`
	Fail     int              `json:"fail" dc:"Fail count"`
	FailList []ImportFailItem `json:"failList" dc:"Fail details"`
}

type ImportFailItem struct {
	Row    int    `json:"row" dc:"Row number"`
	Reason string `json:"reason" dc:"Fail reason"`
}

type ImportTemplateReq struct {
	g.Meta `path:"/user/import-template" method:"get" tags:"User" summary:"Download import template"`
}

type ImportTemplateRes struct{}

type ResetPasswordReq struct {
	g.Meta   `path:"/user/{id}/reset-password" method:"put" tags:"User" summary:"Reset user password"`
	Id       int    `json:"id" v:"required" dc:"User ID"`
	Password string `json:"password" v:"required|length:5,20#请输入密码|密码长度为5-20个字符" dc:"New password"`
}

type ResetPasswordRes struct{}

type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"post" mime:"multipart/form-data" tags:"User" summary:"Upload and update avatar"`
}

type UpdateAvatarRes struct {
	Url string `json:"url" dc:"Avatar URL"`
}

