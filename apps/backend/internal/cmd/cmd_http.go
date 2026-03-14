package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"backend/internal/controller/auth"
	"backend/internal/controller/user"
	"backend/internal/service/middleware"
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

	s.Group("/api", func(group *ghttp.RouterGroup) {
		// Static file serving for uploads (no JSON wrapper)
		group.Group("/uploads", func(group *ghttp.RouterGroup) {
			group.Middleware(middlewareSvc.CORS)
			group.ALL("/*any", func(r *ghttp.Request) {
				basePath := g.Cfg().MustGet(r.Context(), "upload.path", "upload").String()
				pathSuffix := r.GetRouter("any").String()
				filePath := gfile.Join(basePath, pathSuffix)
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
			group.ALLMap(g.Map{
				"POST:/auth/logout": authCtrl.Logout,
				"GET:/auth/codes":   authCtrl.Codes,
			})
			group.Bind(
				user.NewV1(),
			)
		})
	})

	s.Run()
	return
}
