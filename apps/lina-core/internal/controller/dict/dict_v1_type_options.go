package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	"lina-core/internal/model/entity"
)

// TypeOptions 获取字典类型选项列表
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
