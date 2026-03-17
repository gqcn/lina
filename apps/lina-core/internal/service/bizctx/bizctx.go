package bizctx

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/consts"
	"lina-core/internal/model"
)

// Service provides business context operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// Init initializes and injects business context into request.
func (s *Service) Init(r *ghttp.Request, ctx *model.Context) {
	r.SetCtxVar(consts.ContextKey, ctx)
}

// Get retrieves business context from context.
func (s *Service) Get(ctx context.Context) *model.Context {
	value := ctx.Value(consts.ContextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*model.Context); ok {
		return localCtx
	}
	return nil
}

// SetUser sets user info into business context.
func (s *Service) SetUser(ctx context.Context, userId int, username string, status int) {
	if c := s.Get(ctx); c != nil {
		c.UserId = userId
		c.Username = username
		c.Status = status
	}
}
