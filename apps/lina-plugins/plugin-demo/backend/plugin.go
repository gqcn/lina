package backend

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
	democtrl "lina-plugin-demo/backend/internal/controller/demo"
)

const (
	pluginID                = "plugin-demo"
	pluginAfterAuthHeader   = "X-Lina-Plugin-After-Auth"
	pluginPingRoute         = "/plugins/plugin-demo/ping"
	pluginSummaryRoute      = "/plugins/plugin-demo/summary"
	pluginHeartbeatCronName = "plugin-demo-heartbeat"
	pluginHeartbeatCronRule = "# * * * * *"
)

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.RegisterAfterAuthHandler(
		pluginhost.ExtensionPointHTTPRequestAfterAuth,
		pluginhost.CallbackExecutionModeBlocking,
		markAfterAuthRequest,
	)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
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

func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	var (
		middlewares    = registrar.Middlewares()
		demoController = democtrl.NewV1()
	)
	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(demoController.Ping)
		})

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.OperLog(),
			)
			group.Bind(demoController.Summary)
		})
	})
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
