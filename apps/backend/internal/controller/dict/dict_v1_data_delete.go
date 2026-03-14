package dict

import (
	"context"

	v1 "backend/api/dict/v1"
)

func (c *ControllerV1) DataDelete(ctx context.Context, req *v1.DataDeleteReq) (res *v1.DataDeleteRes, err error) {
	err = c.dictSvc.DataDelete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DataDeleteRes{}, nil
}
