package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"backend/internal/model"
	"backend/internal/service/auth"
	"backend/internal/service/bizctx"
)

// Service provides middleware operations.
type Service struct {
	authSvc   *auth.Service
	bizCtxSvc *bizctx.Service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		authSvc:   auth.New(),
		bizCtxSvc: bizctx.New(),
	}
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

	// Inject user info into business context
	s.bizCtxSvc.SetUser(r.Context(), claims.UserId, claims.Username, claims.Status)
	r.Middleware.Next()
}
