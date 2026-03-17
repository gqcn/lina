package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta         `path:"/user" method:"get" tags:"用户管理" summary:"获取用户列表"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数"`
	Username       string `json:"username" dc:"按用户名筛选"`
	Nickname       string `json:"nickname" dc:"按昵称筛选"`
	Status         *int   `json:"status" dc:"按状态筛选"`
	Phone          string `json:"phone" dc:"按手机号筛选"`
	Sex            *int   `json:"sex" dc:"按性别筛选"`
	DeptId         *int   `json:"deptId" dc:"按部门ID筛选"`
	BeginTime      string `json:"beginTime" dc:"按创建时间起始筛选"`
	EndTime        string `json:"endTime" dc:"按创建时间结束筛选"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,username,nickname,phone,email,status,created_at"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc"`
}

type ListItem struct {
	*entity.SysUser
	DeptId   int    `json:"deptId" dc:"部门ID"`
	DeptName string `json:"deptName" dc:"部门名称"`
}

type ListRes struct {
	List  []*ListItem `json:"list" dc:"用户列表"`
	Total int         `json:"total" dc:"总条数"`
}

type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"用户管理" summary:"创建用户"`
	Username string `json:"username" v:"required|length:2,64#请输入用户名|用户名长度为2-64个字符"`
	Password string `json:"password" v:"required|length:6,32#请输入密码|密码长度为6-32个字符"`
	Nickname string `json:"nickname" dc:"昵称"`
	Email    string `json:"email" dc:"邮箱"`
	Phone    string `json:"phone" dc:"手机号"`
	Sex      *int   `json:"sex" d:"0" dc:"性别：0=未知 1=男 2=女"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用"`
	Remark   string `json:"remark" dc:"备注"`
	DeptId   *int   `json:"deptId" dc:"部门ID"`
	PostIds  []int  `json:"postIds" dc:"岗位ID列表"`
}

type CreateRes struct {
	Id int `json:"id" dc:"用户ID"`
}

type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"用户管理" summary:"获取用户详情"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
}

type GetRes struct {
	*entity.SysUser `dc:"用户信息"`
	DeptId          int    `json:"deptId" dc:"部门ID"`
	DeptName        string `json:"deptName" dc:"部门名称"`
	PostIds         []int  `json:"postIds" dc:"岗位ID列表"`
}

type UpdateReq struct {
	g.Meta   `path:"/user/{id}" method:"put" tags:"用户管理" summary:"更新用户"`
	Id       int     `json:"id" v:"required" dc:"用户ID"`
	Username *string `json:"username" dc:"用户名"`
	Password *string `json:"password" dc:"密码（为空则不修改）"`
	Nickname *string `json:"nickname" dc:"昵称"`
	Email    *string `json:"email" dc:"邮箱"`
	Phone    *string `json:"phone" dc:"手机号"`
	Sex      *int    `json:"sex" dc:"性别"`
	Status   *int    `json:"status" dc:"状态"`
	Remark   *string `json:"remark" dc:"备注"`
	DeptId   *int    `json:"deptId" dc:"部门ID"`
	PostIds  []int   `json:"postIds" dc:"岗位ID列表"`
}

type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/user/{id}" method:"delete" tags:"用户管理" summary:"删除用户"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
}

type DeleteRes struct{}

type UpdateStatusReq struct {
	g.Meta `path:"/user/{id}/status" method:"put" tags:"用户管理" summary:"更新用户状态"`
	Id     int `json:"id" v:"required" dc:"用户ID"`
	Status int `json:"status" v:"in:0,1#状态值无效" dc:"状态：1=正常 0=停用"`
}

type UpdateStatusRes struct{}

type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"用户管理" summary:"获取当前用户信息"`
}

type GetProfileRes struct {
	*entity.SysUser `dc:"用户信息"`
}

type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"用户管理" summary:"更新当前用户信息"`
	Nickname *string `json:"nickname" dc:"昵称"`
	Email    *string `json:"email" dc:"邮箱"`
	Phone    *string `json:"phone" dc:"手机号"`
	Sex      *int    `json:"sex" dc:"性别"`
	Password *string `json:"password" dc:"新密码"`
}

type UpdateProfileRes struct{}

type GetInfoReq struct {
	g.Meta `path:"/user/info" method:"get" tags:"用户管理" summary:"获取前端用户信息"`
}

type GetInfoRes struct {
	UserId   int      `json:"userId" dc:"用户ID"`
	Username string   `json:"username" dc:"用户名"`
	RealName string   `json:"realName" dc:"真实姓名"`
	Avatar   string   `json:"avatar" dc:"头像地址"`
	Roles    []string `json:"roles" dc:"用户角色"`
	HomePath string   `json:"homePath" dc:"首页路径"`
}

type ExportReq struct {
	g.Meta `path:"/user/export" method:"get" tags:"用户管理" summary:"导出用户数据" operLog:"4"`
	Ids    []int `json:"ids" dc:"导出指定用户ID列表"`
}

type ExportRes struct{}

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

type ResetPasswordReq struct {
	g.Meta   `path:"/user/{id}/reset-password" method:"put" tags:"用户管理" summary:"重置用户密码"`
	Id       int    `json:"id" v:"required" dc:"用户ID"`
	Password string `json:"password" v:"required|length:5,20#请输入密码|密码长度为5-20个字符" dc:"新密码"`
}

type ResetPasswordRes struct{}

type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"post" mime:"multipart/form-data" tags:"用户管理" summary:"上传并更新头像"`
}

type UpdateAvatarRes struct {
	Url string `json:"url" dc:"头像地址"`
}

type DeptTreeReq struct {
	g.Meta `path:"/user/dept-tree" method:"get" tags:"用户管理" summary:"获取用户筛选部门树"`
}

type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID"`
	Label     string          `json:"label" dc:"部门名称"`
	UserCount int             `json:"userCount" dc:"部门用户数"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表"`
}

type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树"`
}
