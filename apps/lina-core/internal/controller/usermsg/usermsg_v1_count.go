package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

func (c *ControllerV1) Count(ctx context.Context, req *v1.CountReq) (res *v1.CountRes, err error) {
	count, err := c.usermsgSvc.UnreadCount(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CountRes{Count: count}, nil
}
