package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
	runtimectrl "lina-plugin-demo-runtime/backend/internal/controller/runtime"
)

const (
	// ReviewExamplePluginID identifies the runtime sample plugin used by review documents.
	ReviewExamplePluginID = "plugin-demo-runtime"
)

// BuildReviewExampleSourcePlugin returns a review-only source-plugin registration example.
//
// The current host already executes runtime backend contracts embedded into wasm,
// but it still does not dynamically execute runtime plugin Go code. This function
// is therefore intentionally not called from an init hook. Its only purpose is to
// show reviewers how a runtime plugin can keep backend Go sources organized with
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
	var runtimeController = runtimectrl.NewV1()

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
			group.Bind(runtimeController.ReviewSummary)
		})
	})
	return nil
}
