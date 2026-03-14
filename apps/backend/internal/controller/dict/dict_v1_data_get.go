package dict

import (
	"context"

	v1 "backend/api/dict/v1"
)

func (c *ControllerV1) DataGet(ctx context.Context, req *v1.DataGetReq) (res *v1.DataGetRes, err error) {
	dictData, err := c.dictSvc.DataGetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DataGetRes{SysDictData: dictData}, nil
}
