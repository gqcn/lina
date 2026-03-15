package loginlog

import (
	"bytes"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"backend/internal/dao"
	"backend/internal/model/do"
	"backend/internal/model/entity"
)

// Service provides login log operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// CreateInput defines input for Create function.
type CreateInput struct {
	UserName string
	Status   int
	Ip       string
	Browser  string
	Os       string
	Msg      string
}

// Create inserts a new login log record.
func (s *Service) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.SysLoginLog.Ctx(ctx).Data(do.SysLoginLog{
		UserName:  in.UserName,
		Status:    in.Status,
		Ip:        in.Ip,
		Browser:   in.Browser,
		Os:        in.Os,
		Msg:       in.Msg,
		LoginTime: gtime.Now(),
	}).Insert()
	return err
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int
	PageSize       int
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysLoginLog
	Total int
}

// List queries login log list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	if in.UserName != "" {
		m = m.WhereLike(cols.UserName, "%"+in.UserName+"%")
	}
	if in.Ip != "" {
		m = m.WhereLike(cols.Ip, "%"+in.Ip+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.LoginTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.LoginTime, endTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Sorting
	orderBy := cols.LoginTime
	allowedSortFields := map[string]string{
		"id":        cols.Id,
		"loginTime": cols.LoginTime,
	}
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}
	direction := "DESC"
	if in.OrderDirection == "asc" {
		direction = "ASC"
	}

	var list []*entity.SysLoginLog
	err = m.Page(in.PageNum, in.PageSize).
		Order(orderBy + " " + direction).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves login log by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysLoginLog, error) {
	var record *entity.SysLoginLog
	err := dao.SysLoginLog.Ctx(ctx).
		Where(do.SysLoginLog{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("登录日志不存在")
	}
	return record, nil
}

// CleanInput defines input for Clean function.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// Clean hard-deletes login logs by time range.
func (s *Service) Clean(ctx context.Context, in CleanInput) (int, error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	hasFilter := false
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.LoginTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.LoginTime, endTime)
		hasFilter = true
	}
	if !hasFilter {
		m = m.Where(1)
	}

	result, err := m.Delete()
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// DeleteByIds hard-deletes login logs by IDs.
func (s *Service) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.SysLoginLog.Ctx(ctx).WhereIn(dao.SysLoginLog.Columns().Id, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ExportInput defines input for Export function.
type ExportInput struct {
	UserName       string
	Ip             string
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// Export generates an Excel file with login log data.
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysLoginLog.Columns()
	m := dao.SysLoginLog.Ctx(ctx)

	if in.UserName != "" {
		m = m.WhereLike(cols.UserName, "%"+in.UserName+"%")
	}
	if in.Ip != "" {
		m = m.WhereLike(cols.Ip, "%"+in.Ip+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.LoginTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.LoginTime, endTime)
	}

	orderBy := cols.LoginTime
	direction := "DESC"
	if in.OrderDirection == "asc" {
		direction = "ASC"
	}

	var list []*entity.SysLoginLog
	err := m.Order(orderBy + " " + direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"用户名", "状态", "IP地址", "浏览器", "操作系统", "提示消息", "登录时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, log := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), log.UserName)
		statusText := "成功"
		if log.Status == 1 {
			statusText = "失败"
		}
		f.SetCellValue(sheet, cellName(2, row), statusText)
		f.SetCellValue(sheet, cellName(3, row), log.Ip)
		f.SetCellValue(sheet, cellName(4, row), log.Browser)
		f.SetCellValue(sheet, cellName(5, row), log.Os)
		f.SetCellValue(sheet, cellName(6, row), log.Msg)
		if log.LoginTime != nil {
			f.SetCellValue(sheet, cellName(7, row), log.LoginTime.String())
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
