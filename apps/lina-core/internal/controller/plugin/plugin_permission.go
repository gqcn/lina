// This file centralizes plugin-management permission checks for controller actions.
// Plugin management endpoints currently share the common auth middleware, which
// only establishes login/session context. Action-level permission checks remain
// controller-owned until the host introduces a declarative permission middleware.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/role"
)

// pluginManagementPermission identifies one plugin-management action permission.
type pluginManagementPermission string

const (
	pluginManagementPermissionQuery     pluginManagementPermission = "plugin:query"
	pluginManagementPermissionEnable    pluginManagementPermission = "plugin:enable"
	pluginManagementPermissionDisable   pluginManagementPermission = "plugin:disable"
	pluginManagementPermissionInstall   pluginManagementPermission = "plugin:install"
	pluginManagementPermissionUninstall pluginManagementPermission = "plugin:uninstall"
	pluginManagementPermissionWildcard  pluginManagementPermission = "*:*:*"
)

// requirePermission checks whether the current request user owns the target
// plugin-management permission after auth middleware has already established the
// business context for the request.
func (c *ControllerV1) requirePermission(ctx context.Context, permission pluginManagementPermission) error {
	if c == nil || permission == "" {
		return nil
	}

	businessCtx := c.bizCtxSvc.Get(ctx)
	if businessCtx == nil || businessCtx.UserId <= 0 {
		return gerror.New("未获取到当前登录用户")
	}

	accessContext, err := c.roleSvc.GetUserAccessContext(ctx, businessCtx.UserId)
	if err != nil {
		return err
	}
	if hasPluginManagementPermission(accessContext, permission) {
		return nil
	}

	return gerror.Newf("当前用户缺少插件管理权限: %s", permission)
}

func hasPluginManagementPermission(
	accessContext *role.UserAccessContext,
	permission pluginManagementPermission,
) bool {
	if accessContext == nil {
		return false
	}
	if accessContext.IsSuperAdmin {
		return true
	}

	for _, item := range accessContext.Permissions {
		currentPermission := pluginManagementPermission(item)
		if currentPermission == permission || currentPermission == pluginManagementPermissionWildcard {
			return true
		}
	}
	return false
}
