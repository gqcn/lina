package user

import (
	"context"

	"backend/api/user/v1"
)

func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	user, err := c.userSvc.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetInfoRes{
		UserId:   user.Id,
		Username: user.Username,
		RealName: user.Nickname,
		Avatar:   user.Avatar,
		Roles:    []string{"admin"},
		HomePath: "/analytics",
	}, nil
}
