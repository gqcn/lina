package cmd

import (
	"context"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/controller/auth"
	"lina-core/internal/controller/dept"
	"lina-core/internal/controller/dict"
	"lina-core/internal/controller/loginlog"
	"lina-core/internal/controller/notice"
	"lina-core/internal/controller/operlog"
	"lina-core/internal/controller/post"
	"lina-core/internal/controller/sysinfo"
	"lina-core/internal/controller/user"
	"lina-core/internal/controller/usermsg"
	"lina-core/internal/packed"
	"lina-core/internal/service/middleware"
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
	)

	// Set OpenAPI info for API documentation
	oai := s.GetOpenApi()
	oai.Info.Title = "Lina Admin API"
	oai.Info.Description = "Lina 管理后台系统 RESTful API 接口文档。基于 GoFrame 框架构建，提供用户管理、部门管理、岗位管理、字典管理、通知公告、操作日志、登录日志等功能模块的完整接口。"
	oai.Info.Version = "v0.5.0"

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
				"GET:/auth/codes":   authCtrl.Codes,
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
