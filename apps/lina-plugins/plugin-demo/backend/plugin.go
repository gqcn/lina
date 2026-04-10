package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	plugindemo "lina-plugin-demo"
	"lina-core/pkg/pluginhost"
	democtrl "lina-plugin-demo/backend/internal/controller/demo"
)

const (
	pluginID = "plugin-demo"
)

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(plugindemo.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
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
