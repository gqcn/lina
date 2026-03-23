package job

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/locker"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type Service struct {
	lockerSvc *locker.Service
}

func New() *Service {
	return &Service{
		lockerSvc: locker.New(),
	}
}

type ListInput struct {
	Name     string
	Group    string
	Status   *int
	Page     int
	PageSize int
}

type ListOutput struct {
	Items []*entity.SysJob
	Total int
}

func (s *Service) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	m := dao.SysJob.Ctx(ctx)
	if in.Name != "" {
		m = m.WhereLike(dao.SysJob.Columns().Name, "%"+in.Name+"%")
	}
	if in.Group != "" {
		m = m.Where(dao.SysJob.Columns().Group, in.Group)
	}
	if in.Status != nil {
		m = m.Where(dao.SysJob.Columns().Status, *in.Status)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []*entity.SysJob
	err = m.Page(in.Page, in.PageSize).OrderDesc(dao.SysJob.Columns().Id).Scan(&items)
	if err != nil {
		return nil, err
	}

	return &ListOutput{Items: items, Total: total}, nil
}

type CreateInput struct {
	Name        string
	Group       string
	Command     string
	CronExpr    string
	Description string
	Status      int
	Singleton   int
	MaxTimes    int
	Remark      string
}

func (s *Service) Create(ctx context.Context, in *CreateInput) error {
	user := g.RequestFromCtx(ctx).GetHeader("X-User-Name")
	_, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		Name:        in.Name,
		Group:       in.Group,
		Command:     in.Command,
		CronExpr:    in.CronExpr,
		Description: in.Description,
		Status:      in.Status,
		Singleton:   in.Singleton,
		MaxTimes:    in.MaxTimes,
		ExecTimes:   0,
		IsSystem:    0,
		CreateBy:    user,
		CreateTime:  gtime.Now(),
		Remark:      in.Remark,
	}).Insert()
	return err
}

type UpdateInput struct {
	Id          uint64
	Name        string
	Group       string
	Command     string
	CronExpr    string
	Description string
	Status      int
	Singleton   int
	MaxTimes    int
	Remark      string
}

func (s *Service) Update(ctx context.Context, in *UpdateInput) error {
	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).Where(dao.SysJob.Columns().Id, in.Id).Scan(&job)
	if err != nil {
		return err
	}
	if job == nil {
		return gerror.New("任务不存在")
	}

	user := g.RequestFromCtx(ctx).GetHeader("X-User-Name")
	data := do.SysJob{
		Name:        in.Name,
		Group:       in.Group,
		CronExpr:    in.CronExpr,
		Description: in.Description,
		Status:      in.Status,
		Singleton:   in.Singleton,
		MaxTimes:    in.MaxTimes,
		UpdateBy:    user,
		UpdateTime:  gtime.Now(),
		Remark:      in.Remark,
	}

	if job.IsSystem == 0 {
		data.Command = in.Command
	}

	_, err = dao.SysJob.Ctx(ctx).Data(data).Where(dao.SysJob.Columns().Id, in.Id).Update()
	return err
}

func (s *Service) Delete(ctx context.Context, ids []uint64) error {
	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).WhereIn(dao.SysJob.Columns().Id, ids).Scan(&jobs)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if job.IsSystem == 1 {
			return gerror.Newf("系统任务[%s]不可删除", job.Name)
		}
	}

	_, err = dao.SysJob.Ctx(ctx).WhereIn(dao.SysJob.Columns().Id, ids).Delete()
	return err
}

func (s *Service) UpdateStatus(ctx context.Context, id uint64, status int) error {
	user := g.RequestFromCtx(ctx).GetHeader("X-User-Name")
	_, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		Status:     status,
		UpdateBy:   user,
		UpdateTime: gtime.Now(),
	}).Where(dao.SysJob.Columns().Id, id).Update()
	return err
}

func (s *Service) Run(ctx context.Context, id uint64) error {
	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).Where(dao.SysJob.Columns().Id, id).Scan(&job)
	if err != nil {
		return err
	}
	if job == nil {
		return gerror.New("任务不存在")
	}

	go func() {
		_ = s.Execute(context.Background(), job)
	}()

	return nil
}
