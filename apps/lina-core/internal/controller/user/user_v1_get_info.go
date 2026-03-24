package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// GetInfo returns current logged-in user information
func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	user, err := c.userSvc.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	realName := user.Nickname
	if realName == "" {
		realName = user.Username
	}
	return &v1.GetInfoRes{
		UserId:   user.Id,
		Username: user.Username,
		RealName: realName,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Roles:    []string{"admin"},
		HomePath: "/analytics",
	}, nil
}
