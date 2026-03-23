package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/job"
)

func (c *ControllerV1) JobLogList(ctx context.Context, req *v1.JobLogListReq) (res *v1.JobLogListRes, err error) {
	svc := job.New()
	out, err := svc.LogList(ctx, &job.LogListInput{
		JobName:   req.JobName,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Page:      req.Page,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.JobLogItem, len(out.Items))
	for i, item := range out.Items {
		items[i] = &v1.JobLogItem{
			Id:         item.Id,
			JobId:      item.JobId,
			JobName:    item.JobName,
			JobGroup:   item.JobGroup,
			Command:    item.Command,
			Status:     item.Status,
			StartTime:  item.StartTime.String(),
			EndTime:    item.EndTime.String(),
			Duration:   item.Duration,
			ErrorMsg:   item.ErrorMsg,
			CreateTime: item.CreateTime.String(),
		}
	}

	return &v1.JobLogListRes{Items: items, Total: out.Total}, nil
}
