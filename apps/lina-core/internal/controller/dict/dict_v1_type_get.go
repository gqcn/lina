package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// TypeGet 获取字典类型详情
func (c *ControllerV1) TypeGet(ctx context.Context, req *v1.TypeGetReq) (res *v1.TypeGetRes, err error) {
	dictType, err := c.dictSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TypeGetRes{SysDictType: dictType}, nil
}
