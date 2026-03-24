package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// Get returns user details
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	user, err := c.userSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	deptId, deptName, _ := c.userSvc.GetUserDeptInfo(ctx, req.Id)
	postIds, _ := c.userSvc.GetUserPostIds(ctx, req.Id)
	if postIds == nil {
		postIds = []int{}
	}
	return &v1.GetRes{
		SysUser:  user,
		DeptId:   deptId,
		DeptName: deptName,
		PostIds:  postIds,
	}, nil
}
