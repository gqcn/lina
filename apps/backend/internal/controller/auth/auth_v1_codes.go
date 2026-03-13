package auth

import (
	"context"

	"backend/api/auth/v1"
)

func (c *ControllerV1) Codes(ctx context.Context, req *v1.CodesReq) (res *v1.CodesRes, err error) {
	// No RBAC yet, return empty access codes
	return &v1.CodesRes{Codes: []string{}}, nil
}
