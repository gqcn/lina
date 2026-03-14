package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"backend/internal/controller/auth"
	"backend/internal/controller/user"
	"backend/internal/service/middleware"
)

// initDatabase executes SQL init and seed files if they exist.
func initDatabase(ctx context.Context) {
	sqlFiles := g.Cfg().MustGet(ctx, "init.sqlFiles").Strings()
	if len(sqlFiles) == 0 {
		g.Log().Warning(ctx, "no SQL files configured in init.sqlFiles")
		return
	}
	for _, file := range sqlFiles {
		if !gfile.Exists(file) {
			continue
		}
		sql := gfile.GetContents(file)
		if sql == "" {
			continue
		}
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			g.Log().Warningf(ctx, "execute %s: %v", file, err)
		}
	}
}

var (
	Init = gcmd.Command{
		Name:  "init",
		Usage: "init",
		Brief: "initialize database (run SQL init/seed scripts)",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initDatabase(ctx)
			g.Log().Info(ctx, "Database initialization completed.")
			return nil
		},
	}

	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start lina backend server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {

			var (
				s             = g.Server()
				middlewareSvc = middleware.New()
			)

			var authCtrl = auth.NewV1()

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
			return nil
		},
	}
)
