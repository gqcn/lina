package pluginhost

import "github.com/gogf/gf/v2/net/ghttp"

// PluginEnabledChecker defines one host callback that reports whether a plugin is currently enabled.
type PluginEnabledChecker func(pluginID string) bool

// RouteHandler defines one plugin-owned HTTP handler.
type RouteHandler = ghttp.HandlerFunc

// RouteRegistrars exposes public and protected route registrars for one plugin.
type RouteRegistrars interface {
	// Public returns the public route registrar.
	Public() RouteGroupRegistrar
	// Protected returns the protected route registrar.
	Protected() RouteGroupRegistrar
}

// RouteGroupRegistrar exposes guarded route registration methods.
type RouteGroupRegistrar interface {
	// Bind registers one or more guarded handler objects using GoFrame object routing.
	Bind(handlerOrObject ...interface{})
	// GET registers one guarded GET route.
	GET(pattern string, handler RouteHandler)
	// POST registers one guarded POST route.
	POST(pattern string, handler RouteHandler)
	// PUT registers one guarded PUT route.
	PUT(pattern string, handler RouteHandler)
	// DELETE registers one guarded DELETE route.
	DELETE(pattern string, handler RouteHandler)
	// ALL registers one guarded all-method route.
	ALL(pattern string, handler RouteHandler)
}

type routeRegistrars struct {
	public    RouteGroupRegistrar
	protected RouteGroupRegistrar
}

type routeGroup struct {
	group          *ghttp.RouterGroup
	pluginID       string
	enabledChecker PluginEnabledChecker
}

// NewRouteRegistrars creates and returns a new RouteRegistrars instance.
func NewRouteRegistrars(
	publicGroup *ghttp.RouterGroup,
	protectedGroup *ghttp.RouterGroup,
	pluginID string,
	enabledChecker PluginEnabledChecker,
) RouteRegistrars {
	return &routeRegistrars{
		public: &routeGroup{
			group:          publicGroup,
			pluginID:       pluginID,
			enabledChecker: enabledChecker,
		},
		protected: &routeGroup{
			group:          protectedGroup,
			pluginID:       pluginID,
			enabledChecker: enabledChecker,
		},
	}
}

// Public returns the public route registrar.
func (r *routeRegistrars) Public() RouteGroupRegistrar {
	if r == nil {
		return nil
	}
	return r.public
}

// Protected returns the protected route registrar.
func (r *routeRegistrars) Protected() RouteGroupRegistrar {
	if r == nil {
		return nil
	}
	return r.protected
}

// GET registers one guarded GET route.
func (r *routeGroup) GET(pattern string, handler RouteHandler) {
	if r == nil || r.group == nil {
		return
	}
	r.group.GET(pattern, r.wrap(handler))
}

// POST registers one guarded POST route.
func (r *routeGroup) POST(pattern string, handler RouteHandler) {
	if r == nil || r.group == nil {
		return
	}
	r.group.POST(pattern, r.wrap(handler))
}

// PUT registers one guarded PUT route.
func (r *routeGroup) PUT(pattern string, handler RouteHandler) {
	if r == nil || r.group == nil {
		return
	}
	r.group.PUT(pattern, r.wrap(handler))
}

// DELETE registers one guarded DELETE route.
func (r *routeGroup) DELETE(pattern string, handler RouteHandler) {
	if r == nil || r.group == nil {
		return
	}
	r.group.DELETE(pattern, r.wrap(handler))
}

// ALL registers one guarded all-method route.
func (r *routeGroup) ALL(pattern string, handler RouteHandler) {
	if r == nil || r.group == nil {
		return
	}
	r.group.ALL(pattern, r.wrap(handler))
}

func (r *routeGroup) wrap(handler RouteHandler) RouteHandler {
	if handler == nil {
		panic("pluginhost: route handler is nil")
	}
	return func(req *ghttp.Request) {
		if !r.allow(req) {
			return
		}
		handler(req)
	}
}

// Bind registers one or more guarded handler objects using GoFrame object routing.
func (r *routeGroup) Bind(handlerOrObject ...interface{}) {
	if r == nil || r.group == nil || len(handlerOrObject) == 0 {
		return
	}

	r.group.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(func(req *ghttp.Request) {
			if !r.allow(req) {
				return
			}
			req.Middleware.Next()
		})
		group.Bind(handlerOrObject...)
	})
}

func (r *routeGroup) allow(req *ghttp.Request) bool {
	if req == nil {
		return false
	}
	if r.enabledChecker != nil && !r.enabledChecker(r.pluginID) {
		req.Response.WriteStatus(404)
		req.ExitAll()
		return false
	}
	return true
}
