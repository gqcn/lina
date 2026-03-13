package user

import (
	"context"

	"backend/api/user/v1"
	usersvc "backend/internal/service/user"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	id, err := c.userSvc.Create(ctx, usersvc.CreateInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   status,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
