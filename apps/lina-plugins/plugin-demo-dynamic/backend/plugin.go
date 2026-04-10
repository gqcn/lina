package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
	dynamicctrl "lina-plugin-demo-dynamic/backend/internal/controller/dynamic"
)

const (
	// ReviewExamplePluginID identifies the dynamic sample plugin used by review documents.
	ReviewExamplePluginID = "plugin-demo-dynamic"
)

// BuildReviewExampleSourcePlugin returns a review-only source-plugin registration example.
//
// The current host already executes dynamic backend contracts embedded into wasm,
// but it still does not dynamically execute plugin-owned Go code for dynamic plugins. This function
// is therefore intentionally not called from an init hook. Its only purpose is to
// show reviewers how a dynamic plugin can keep backend Go sources organized with
// the same directory structure as a source plugin while the executable runtime ABI
// stays constrained to the host-owned declarative contract.
func BuildReviewExampleSourcePlugin() *pluginhost.SourcePlugin {
	plugin := pluginhost.NewSourcePlugin(ReviewExamplePluginID)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	return plugin
}

func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	var dynamicController = dynamicctrl.NewV1()

	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		middlewares := registrar.Middlewares()
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
			middlewares.Auth(),
			middlewares.OperLog(),
		)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(dynamicController.ReviewSummary)
		})
	})
	return nil
}
