package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/operlog"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/session"
	"lina-core/pkg/pluginhost"
)

// Service provides middleware operations.
type Service struct {
	authSvc    *auth.Service      // Authentication service
	bizCtxSvc  *bizctx.Service    // Business context service
	operLogSvc *operlog.Service   // Operation log service
	pluginSvc  *pluginsvc.Service // Plugin service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		authSvc:    auth.New(),
		bizCtxSvc:  bizctx.New(),
		operLogSvc: operlog.New(),
		pluginSvc:  pluginsvc.New(),
	}
}

// SessionStore returns the session store for external use (e.g., cleanup tasks).
func (s *Service) SessionStore() session.Store {
	return s.authSvc.SessionStore()
}

// Ctx injects business context into request.
func (s *Service) Ctx(r *ghttp.Request) {
	customCtx := &model.Context{}
	s.bizCtxSvc.Init(r, customCtx)
	r.Middleware.Next()
}

// CORS handles cross-origin requests.
func (s *Service) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Auth validates JWT token and injects user info into context.
func (s *Service) Auth(r *ghttp.Request) {
	tokenHeader := r.GetHeader("Authorization")
	if tokenHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if tokenString == tokenHeader {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	claims, err := s.authSvc.ParseToken(r.Context(), tokenString)
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Update last active time and validate session exists (supports forced logout and timeout cleanup)
	exists, err := s.authSvc.SessionStore().TouchOrValidate(r.Context(), claims.TokenId)
	if err != nil || !exists {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Inject user info into business context
	s.bizCtxSvc.SetUser(r.Context(), claims.TokenId, claims.UserId, claims.Username, claims.Status)
	s.pluginSvc.DispatchAfterAuthRequest(
		r.Context(),
		pluginhost.NewAfterAuthInput(
			r,
			claims.TokenId,
			claims.UserId,
			claims.Username,
			claims.Status,
		),
	)
	r.Middleware.Next()
}
