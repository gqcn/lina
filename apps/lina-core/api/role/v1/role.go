package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ControllerV1 is the role controller interface.
type ControllerV1 struct {
	g.Meta `group:"role" prefix:"/api/v1"`
}

// Role module API endpoints:
// - GET    /role          -> List (分页查询角色列表)
// - GET    /role/:id      -> Get (查询角色详情)
// - POST   /role          -> Create (创建角色)
// - PUT    /role/:id      -> Update (更新角色)
// - DELETE /role/:id      -> Delete (删除角色)
// - PUT    /role/:id/status -> Status (切换角色状态)
// - GET    /role/options  -> Options (查询角色下拉选项)
// - GET    /role/:id/users -> Users (查询角色用户列表)
// - POST   /role/:id/users -> AssignUsers (分配用户到角色)
// - DELETE /role/:id/users/:userId -> UnassignUser (取消用户授权)