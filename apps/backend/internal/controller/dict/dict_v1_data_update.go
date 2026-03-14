package dict

import (
	"context"

	v1 "backend/api/dict/v1"
	dictsvc "backend/internal/service/dict"
)

func (c *ControllerV1) DataUpdate(ctx context.Context, req *v1.DataUpdateReq) (res *v1.DataUpdateRes, err error) {
	err = c.dictSvc.DataUpdate(ctx, dictsvc.DataUpdateInput{
		Id:       req.Id,
		DictType: req.DictType,
		Label:    req.Label,
		Value:    req.Value,
		Sort:     req.Sort,
		TagStyle: req.TagStyle,
		CssClass: req.CssClass,
		Status:   req.Status,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DataUpdateRes{}, nil
}
