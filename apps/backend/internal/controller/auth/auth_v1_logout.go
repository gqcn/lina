package auth

import (
	"context"

	"backend/api/auth/v1"
)

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// JWT is stateless, logout is handled on frontend by clearing token.
	return &v1.LogoutRes{}, nil
}
