package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
)

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// Record logout log
	if bizCtx := c.bizCtxSvc.Get(ctx); bizCtx != nil {
		c.authSvc.Logout(ctx, bizCtx.Username)
	}
	// JWT is stateless, logout is handled on frontend by clearing token.
	return &v1.LogoutRes{}, nil
}
