package dict

import (
	"context"

	v1 "backend/api/dict/v1"
	dictsvc "backend/internal/service/dict"
)

func (c *ControllerV1) TypeUpdate(ctx context.Context, req *v1.TypeUpdateReq) (res *v1.TypeUpdateRes, err error) {
	err = c.dictSvc.Update(ctx, dictsvc.UpdateInput{
		Id:     req.Id,
		Name:   req.Name,
		Type:   req.Type,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TypeUpdateRes{}, nil
}
