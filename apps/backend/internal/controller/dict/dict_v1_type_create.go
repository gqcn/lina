package dict

import (
	"context"

	v1 "backend/api/dict/v1"
	dictsvc "backend/internal/service/dict"
)

func (c *ControllerV1) TypeCreate(ctx context.Context, req *v1.TypeCreateReq) (res *v1.TypeCreateRes, err error) {
	id, err := c.dictSvc.Create(ctx, dictsvc.CreateInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: *req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TypeCreateRes{Id: id}, nil
}
