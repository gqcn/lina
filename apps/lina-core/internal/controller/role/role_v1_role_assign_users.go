package role

import (
	"context"

	"lina-core/api/role/v1"
)

func (c *ControllerV1) RoleAssignUsers(ctx context.Context, req *v1.RoleAssignUsersReq) (res *v1.RoleAssignUsersRes, err error) {
	// Assign users to role
	err = c.roleSvc.AssignUsers(ctx, req.Id, req.UserIds)
	if err != nil {
		return nil, err
	}

	return &v1.RoleAssignUsersRes{}, nil
}