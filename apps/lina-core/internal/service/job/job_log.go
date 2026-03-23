package job

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

type LogListInput struct {
	JobName   string
	Status    *int
	StartTime string
	EndTime   string
	Page      int
	PageSize  int
}

type LogListOutput struct {
	Items []*entity.SysJobLog
	Total int
}

func (s *Service) LogList(ctx context.Context, in *LogListInput) (*LogListOutput, error) {
	m := dao.SysJobLog.Ctx(ctx)
	if in.JobName != "" {
		m = m.WhereLike(dao.SysJobLog.Columns().JobName, "%"+in.JobName+"%")
	}
	if in.Status != nil {
		m = m.Where(dao.SysJobLog.Columns().Status, *in.Status)
	}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.SysJobLog.Columns().StartTime, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.SysJobLog.Columns().StartTime, in.EndTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []*entity.SysJobLog
	err = m.Page(in.Page, in.PageSize).OrderDesc(dao.SysJobLog.Columns().Id).Scan(&items)
	if err != nil {
		return nil, err
	}

	return &LogListOutput{Items: items, Total: total}, nil
}
