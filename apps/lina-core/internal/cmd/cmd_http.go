package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/controller/auth"
	configctrl "lina-core/internal/controller/config"
	"lina-core/internal/controller/dept"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
	"lina-core/internal/controller/loginlog"
	"lina-core/internal/controller/menu"
	monitorctrl "lina-core/internal/controller/monitor"
	"lina-core/internal/controller/notice"
	"lina-core/internal/controller/operlog"
	pluginctrl "lina-core/internal/controller/plugin"
	"lina-core/internal/controller/post"
	"lina-core/internal/controller/role"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/packed"
	"lina-core/internal/service/config"
	"lina-core/internal/service/cron"
	"lina-core/internal/service/election"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
)

type HttpInput struct {
	g.Meta `name:"http" brief:"start http server"`
}
type HttpOutput struct{}

func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	var (
		s             = g.Server()
		configSvc     = config.New()
		middlewareSvc = middleware.New()
		authCtrl      = auth.NewV1()
	)

	// Initialize distributed locker and leader election
	var (
		lockerSvc   = locker.New()
		electionCfg = configSvc.GetElection(ctx)
		electionSvc = election.New(lockerSvc, electionCfg)
		sessionCfg  = configSvc.GetSession(ctx)
		monCfg      = configSvc.GetMonitor(ctx)
		cronSvc     = cron.New(sessionCfg, monCfg, middlewareSvc.SessionStore(), electionSvc)
	)
	// Start election when distributed deployment.
	electionSvc.Start(ctx)
	// Start all cron jobs (session cleanup, server monitor, etc.)
	cronSvc.Start(ctx)

	// Enhance OpenAPI documentation with config values and JWT security scheme.
	m.enhanceOpenAPIDocs(ctx, s, configSvc)

	// =============================================================================================
	// Dynamic routes registering.
	// =============================================================================================

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			ghttp.MiddlewareHandlerResponse,
			middlewareSvc.CORS,
			middlewareSvc.Ctx,
		)

		// Static file serving for uploads.
		group.Group("/uploads", func(group *ghttp.RouterGroup) {
			group.ALL("/*any", func(r *ghttp.Request) {
				var (
					uploadCfg  = configSvc.GetUpload(r.Context())
					pathSuffix = r.GetRouter("any").String()
					filePath   = gfile.Join(uploadCfg.Path, pathSuffix)
				)
				if !gfile.Exists(filePath) {
					r.Response.WriteStatus(404)
					r.ExitAll()
					return
				}
				r.Response.ServeFile(filePath)
				r.ExitAll()
			})
		})

		// Public routes (no auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(authCtrl.Login)
		})

		// Protected routes (auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middlewareSvc.Auth,
				middlewareSvc.OperLog,
			)
			group.Bind(
				authCtrl.Logout,
				user.NewV1(),
				dict.NewV1(),
				dept.NewV1(),
				post.NewV1(),
				menu.NewV1(),
				role.NewV1(),
				notice.NewV1(),
				usermsg.NewV1(),
				loginlog.NewV1(),
				operlog.NewV1(),
				sysinfo.NewV1(),
				filectrl.NewV1(),
				monitorctrl.NewV1(),
				configctrl.NewV1(),
				pluginctrl.NewV1(),
			)
		})
	})

	// =============================================================================================
	// Static service for frontend assets.
	// =============================================================================================

	// Serve embedded frontend static files
	subFS, _ := fs.Sub(packed.Files, "public")
	fileServer := http.FileServer(http.FS(subFS))
	s.BindHandler("/*", func(r *ghttp.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := subFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
			r.ExitAll()
			return
		}
		// SPA fallback: serve index.html for unmatched paths
		r.Request.URL.Path = "/index.html"
		fileServer.ServeHTTP(r.Response.RawWriter(), r.Request)
		r.ExitAll()
	})

	if err = pluginsvc.New().DispatchHookEvent(ctx, pluginsvc.HookEventSystemStarted, map[string]interface{}{}); err != nil {
		g.Log().Warningf(ctx, "dispatch system.started plugin hook failed: %v", err)
	}

	s.Run()
	return
}

func (m *Main) enhanceOpenAPIDocs(
	ctx context.Context,
	server *ghttp.Server,
	configSvc *config.Service,
) {
	// Set OpenAPI info from configuration
	oaiCfg := configSvc.GetOpenApi(ctx)
	oai := server.GetOpenApi()
	oai.Info.Title = oaiCfg.Title
	oai.Info.Description = oaiCfg.Description
	oai.Info.Version = oaiCfg.Version
	oai.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	oai.Config.CommonResponseDataField = "Data"

	// Set API server URL so documentation shows the correct backend address
	if oaiCfg.ServerUrl != "" {
		oai.Servers = &goai.Servers{
			{
				URL:         oaiCfg.ServerUrl,
				Description: oaiCfg.ServerDescription,
			},
		}
	}

	// Add JWT Bearer security scheme for API documentation
	oai.Components.SecuritySchemes = goai.SecuritySchemes{
		"BearerAuth": goai.SecuritySchemeRef{
			Value: &goai.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT Bearer Token Authentication",
				In:           "header",
				Name:         "Authorization",
			},
		},
	}
	oai.Security = &goai.SecurityRequirements{
		{"BearerAuth": {}},
	}
}
