package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobList(ctx context.Context, req *v1.JobListReq) (res *v1.JobListRes, err error) {
	svc := job.New()
	out, err := svc.List(ctx, &job.ListInput{
		Name:     req.Name,
		Group:    req.Group,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.JobItem, len(out.Items))
	for i, item := range out.Items {
		items[i] = &v1.JobItem{
			Id:          item.Id,
			Name:        item.Name,
			Group:       item.Group,
			Command:     item.Command,
			CronExpr:    item.CronExpr,
			Description: item.Description,
			Status:      item.Status,
			Singleton:   item.Singleton,
			MaxTimes:    item.MaxTimes,
			ExecTimes:   item.ExecTimes,
			IsSystem:    item.IsSystem,
			CreateBy:    item.CreateBy,
			CreateTime:  item.CreateTime.String(),
			UpdateBy:    item.UpdateBy,
			UpdateTime:  item.UpdateTime.String(),
			Remark:      item.Remark,
		}
	}

	return &v1.JobListRes{Items: items, Total: out.Total}, nil
}
