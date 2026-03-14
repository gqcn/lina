package dict

import (
	"context"

	v1 "backend/api/dict/v1"
)

func (c *ControllerV1) TypeDelete(ctx context.Context, req *v1.TypeDeleteReq) (res *v1.TypeDeleteRes, err error) {
	err = c.dictSvc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TypeDeleteRes{}, nil
}
