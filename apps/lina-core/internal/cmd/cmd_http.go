package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtimer"

	"lina-core/internal/controller/auth"
	"lina-core/internal/controller/dept"
	"lina-core/internal/controller/dict"
	filectrl "lina-core/internal/controller/file"
	"lina-core/internal/controller/loginlog"
	monitorctrl "lina-core/internal/controller/monitor"
	"lina-core/internal/controller/notice"
	"lina-core/internal/controller/operlog"
	"lina-core/internal/controller/post"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/packed"
	"lina-core/internal/service/middleware"
	"lina-core/internal/service/servermon"
)

type HttpInput struct {
	g.Meta `name:"http" brief:"start http server"`
}
type HttpOutput struct{}

func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	var (
		s             = g.Server()
		middlewareSvc = middleware.New()
		authCtrl      = auth.NewV1()
		serverMonSvc  = servermon.New()
	)

	// Start server monitor collector
	serverMonSvc.StartCollector(ctx)

	// Start session cleanup timer
	cleanupMinute := g.Cfg().MustGet(ctx, "session.cleanupMinute", 5).Int()
	timeoutHour := g.Cfg().MustGet(ctx, "session.timeoutHour", 24).Int()
	sessionStore := middlewareSvc.SessionStore()
	gtimer.Add(ctx, time.Duration(cleanupMinute)*time.Minute, func(ctx context.Context) {
		cleaned, err := sessionStore.CleanupInactive(ctx, timeoutHour)
		if err != nil {
			g.Log().Warningf(ctx, "session cleanup error: %v", err)
		} else if cleaned > 0 {
			g.Log().Infof(ctx, "session cleanup: removed %d inactive sessions", cleaned)
		}
	})

	// Set OpenAPI info from configuration
	oai := s.GetOpenApi()
	oai.Info.Title = g.Cfg().MustGet(ctx, "openapi.title", "Lina Admin API").String()
	oai.Info.Description = g.Cfg().MustGet(ctx, "openapi.description").String()
	oai.Info.Version = g.Cfg().MustGet(ctx, "openapi.version", "v1.0.0").String()

	// Set API server URL so documentation shows the correct backend address
	if serverUrl := g.Cfg().MustGet(ctx, "openapi.serverUrl").String(); serverUrl != "" {
		oai.Servers = &goai.Servers{
			{
				URL:         serverUrl,
				Description: g.Cfg().MustGet(ctx, "openapi.serverDescription", "API Server").String(),
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
				Description:  "JWT Bearer Token 认证",
			},
		},
	}
	oai.Security = &goai.SecurityRequirements{
		{"BearerAuth": {}},
	}

	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// Static file serving for uploads (no JSON wrapper)
		group.Group("/uploads", func(group *ghttp.RouterGroup) {
			group.Middleware(middlewareSvc.CORS)
			group.ALL("/*any", func(r *ghttp.Request) {
				var (
					basePath   = g.Cfg().MustGet(r.Context(), "upload.path", "upload").String()
					pathSuffix = r.GetRouter("any").String()
					filePath   = gfile.Join(basePath, pathSuffix)
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

		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			ghttp.MiddlewareHandlerResponse,
			middlewareSvc.CORS,
			middlewareSvc.Ctx,
		)

		// Public routes (no auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.ALLMap(g.Map{
				"POST:/auth/login": authCtrl.Login,
			})
		})

		// Protected routes (auth required)
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middlewareSvc.Auth)
			group.Middleware(middlewareSvc.OperLog)
			group.ALLMap(g.Map{
				"POST:/auth/logout": authCtrl.Logout,
			})
			group.Bind(
				user.NewV1(),
				dict.NewV1(),
				dept.NewV1(),
				post.NewV1(),
				notice.NewV1(),
				usermsg.NewV1(),
				loginlog.NewV1(),
				operlog.NewV1(),
				sysinfo.NewV1(),
				filectrl.NewV1(),
				monitorctrl.NewV1(),
			)
		})
	})

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

	s.Run()
	return
}
