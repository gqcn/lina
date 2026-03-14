package dict

import (
	"context"

	v1 "backend/api/dict/v1"
	"backend/internal/model/entity"
)

func (c *ControllerV1) TypeOptions(ctx context.Context, req *v1.TypeOptionsReq) (res *v1.TypeOptionsRes, err error) {
	options, err := c.dictSvc.Options(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*entity.SysDictType, 0, len(options))
	for _, opt := range options {
		list = append(list, &entity.SysDictType{
			Id:   opt.Id,
			Name: opt.Name,
			Type: opt.Type,
		})
	}
	return &v1.TypeOptionsRes{List: list}, nil
}
