package user

import (
	"context"

	v1 "backend/api/user/v1"
	usersvc "backend/internal/service/user"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	return nil, c.userSvc.Update(ctx, usersvc.UpdateInput{
		Id:       req.Id,
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Sex:      req.Sex,
		Status:   req.Status,
		Remark:   req.Remark,
		DeptId:   req.DeptId,
		PostIds:  req.PostIds,
	})
}
