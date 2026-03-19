package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	out, err := c.fileSvc.Detail(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	usageItems := make([]*v1.DetailUsageItem, len(out.UsageScenes))
	for i, item := range out.UsageScenes {
		usageItems[i] = &v1.DetailUsageItem{
			Scene:     item.Scene,
			Label:     item.Label,
			BizId:     item.BizId,
			CreatedAt: item.CreatedAt,
		}
	}
	return &v1.DetailRes{
		SysFile:       out.SysFile,
		CreatedByName: out.CreatedByName,
		UsageScenes:   usageItems,
	}, nil
}
