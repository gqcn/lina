package user

import (
	"context"

	"backend/api/user/v1"
	usersvc "backend/internal/service/user"
)

func (c *ControllerV1) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error) {
	return nil, c.userSvc.UpdateProfile(ctx, usersvc.UpdateProfileInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
}
