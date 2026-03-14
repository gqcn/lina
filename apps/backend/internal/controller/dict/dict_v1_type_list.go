package dict

import (
	"context"

	v1 "backend/api/dict/v1"
	dictsvc "backend/internal/service/dict"
)

func (c *ControllerV1) TypeList(ctx context.Context, req *v1.TypeListReq) (res *v1.TypeListRes, err error) {
	out, err := c.dictSvc.List(ctx, dictsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Name:     req.Name,
		Type:     req.Type,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TypeListRes{List: out.List, Total: out.Total}, nil
}
