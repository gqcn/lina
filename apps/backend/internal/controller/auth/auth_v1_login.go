package auth

import (
	"context"

	"backend/api/auth/v1"
	authsvc "backend/internal/service/auth"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	out, err := c.authSvc.Login(ctx, authsvc.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{AccessToken: out.AccessToken}, nil
}
