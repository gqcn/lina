package backend

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/pluginhost"
)

const (
	pluginID                = "plugin-demo"
	pluginAfterAuthHeader   = "X-Lina-Plugin-After-Auth"
	pluginSummaryRoute      = "/plugins/plugin-demo/summary"
	pluginHeartbeatCronName = "plugin-demo-heartbeat"
	pluginHeartbeatCronRule = "# * * * * *"
)

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)

	// plugin-demo 只保留最小后端接入示例，聚焦展示：
	// 1. 插件如何注册宿主路由；
	// 2. 插件如何在鉴权完成后扩展响应；
	// 3. 插件如何注册受宿主启停控制的定时任务。
	plugin.RegisterAfterAuthHandler(
		pluginhost.ExtensionPointHTTPRequestAfterAuth,
		pluginhost.CallbackExecutionModeBlocking,
		markAfterAuthRequest,
	)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerSummaryRoute,
	)
	plugin.RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerHeartbeatCron,
	)

	pluginhost.RegisterSourcePlugin(plugin)
}

// markAfterAuthRequest demonstrates how one plugin mutates the current authenticated response.
func markAfterAuthRequest(ctx context.Context, input pluginhost.AfterAuthInput) error {
	if input == nil {
		return nil
	}

	input.SetResponseHeader(pluginAfterAuthHeader, pluginID)
	return nil
}

// registerSummaryRoute exposes one lightweight protected route so the host can verify plugin route assembly.
func registerSummaryRoute(
	ctx context.Context,
	registrars pluginhost.RouteRegistrars,
) error {
	if registrars == nil || registrars.Protected() == nil {
		return nil
	}

	registrars.Protected().GET(pluginSummaryRoute, func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"pluginId": pluginID,
			"message":  "plugin-demo 仅演示最小源码插件接入，不包含数据库读写示例。",
			"extensionPoints": []string{
				pluginhost.ExtensionPointHTTPRouteRegister.String(),
				pluginhost.ExtensionPointHTTPRequestAfterAuth.String(),
				pluginhost.ExtensionPointCronRegister.String(),
			},
			"callbackModes": []string{
				pluginhost.CallbackExecutionModeBlocking.String(),
			},
			"cronJobName":      pluginHeartbeatCronName,
			"cronPattern":      pluginHeartbeatCronRule,
			"cronPrimaryAware": true,
			"generatedAt":      gtime.Now().String(),
		})
		r.ExitAll()
	})

	g.Log().Infof(ctx, "plugin-demo registered summary route: %s", pluginSummaryRoute)
	return nil
}

// registerHeartbeatCron demonstrates a plugin-owned cron job that can branch on the host primary-node role.
func registerHeartbeatCron(
	ctx context.Context,
	registrar pluginhost.CronRegistrar,
) error {
	if registrar == nil {
		return nil
	}

	return registrar.Add(ctx, pluginHeartbeatCronRule, pluginHeartbeatCronName, func(jobCtx context.Context) error {
		// 这里显式判断主节点角色，示例插件只在主节点打印执行日志，
		// 便于后续其他插件在需要时扩展为“仅主节点执行业务逻辑”的模式。
		if !registrar.IsPrimaryNode() {
			g.Log().Debugf(jobCtx, "plugin-demo heartbeat cron skipped on non-primary node")
			return nil
		}

		g.Log().Debugf(jobCtx, "plugin-demo heartbeat cron executed on primary node")
		return nil
	})
}
