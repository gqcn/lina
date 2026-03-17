package notice

import (
	"context"

	v1 "lina-core/api/notice/v1"
	noticesvc "lina-core/internal/service/notice"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.noticeSvc.Update(ctx, noticesvc.UpdateInput{
		Id:      req.Id,
		Title:   req.Title,
		Type:    req.Type,
		Content: req.Content,
		Status:  req.Status,
		Remark:  req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
