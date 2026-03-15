package operlog

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

// Service provides operation log operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title         string
	OperSummary   string
	OperType      int
	Method        string
	RequestMethod string
	OperName      string
	OperUrl       string
	OperIp        string
	OperParam     string
	JsonResult    string
	Status        int
	ErrorMsg      string
	CostTime      int
}

// Create inserts a new operation log record.
func (s *Service) Create(ctx context.Context, in CreateInput) error {
	_, err := dao.SysOperLog.Ctx(ctx).Data(do.SysOperLog{
		Title:         in.Title,
		OperSummary:   in.OperSummary,
		OperType:      in.OperType,
		Method:        in.Method,
		RequestMethod: in.RequestMethod,
		OperName:      in.OperName,
		OperUrl:       in.OperUrl,
		OperIp:        in.OperIp,
		OperParam:     in.OperParam,
		JsonResult:    in.JsonResult,
		Status:        in.Status,
		ErrorMsg:      in.ErrorMsg,
		CostTime:      in.CostTime,
		OperTime:      gtime.Now(),
	}).Insert()
	return err
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int
	PageSize       int
	Title          string
	OperName       string
	OperType       *int
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysOperLog
	Total int
}

// List queries operation log list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	if in.Title != "" {
		m = m.WhereLike(cols.Title, "%"+in.Title+"%")
	}
	if in.OperName != "" {
		m = m.WhereLike(cols.OperName, "%"+in.OperName+"%")
	}
	if in.OperType != nil {
		m = m.Where(cols.OperType, *in.OperType)
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.OperTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.OperTime, endTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Sorting
	orderBy := cols.OperTime
	allowedSortFields := map[string]string{
		"id":       cols.Id,
		"operTime": cols.OperTime,
		"costTime": cols.CostTime,
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

	var list []*entity.SysOperLog
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

// GetById retrieves operation log by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysOperLog, error) {
	var record *entity.SysOperLog
	err := dao.SysOperLog.Ctx(ctx).
		Where(do.SysOperLog{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("操作日志不存在")
	}
	return record, nil
}

// CleanInput defines input for Clean function.
type CleanInput struct {
	BeginTime string
	EndTime   string
}

// Clean hard-deletes operation logs by time range.
func (s *Service) Clean(ctx context.Context, in CleanInput) (int, error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	hasFilter := false
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.OperTime, in.BeginTime)
		hasFilter = true
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.OperTime, endTime)
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

// DeleteByIds hard-deletes operation logs by IDs.
func (s *Service) DeleteByIds(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := dao.SysOperLog.Ctx(ctx).WhereIn(dao.SysOperLog.Columns().Id, ids).Delete()
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Title          string
	OperName       string
	OperType       *int
	Status         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// Export generates an Excel file with operation log data.
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysOperLog.Columns()
	m := dao.SysOperLog.Ctx(ctx)

	if in.Title != "" {
		m = m.WhereLike(cols.Title, "%"+in.Title+"%")
	}
	if in.OperName != "" {
		m = m.WhereLike(cols.OperName, "%"+in.OperName+"%")
	}
	if in.OperType != nil {
		m = m.Where(cols.OperType, *in.OperType)
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.OperTime, in.BeginTime)
	}
	if in.EndTime != "" {
		endTime := in.EndTime
		if len(endTime) == 10 {
			endTime += " 23:59:59"
		}
		m = m.WhereLTE(cols.OperTime, endTime)
	}

	orderBy := cols.OperTime
	direction := "DESC"
	if in.OrderDirection == "asc" {
		direction = "ASC"
	}

	var list []*entity.SysOperLog
	err := m.Order(orderBy + " " + direction).Scan(&list)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"模块名称", "操作名称", "操作类型", "操作人", "请求方式", "请求URL", "操作IP", "请求参数", "响应结果", "状态", "错误信息", "耗时(ms)", "操作时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	operTypeMap := map[int]string{
		1: "新增", 2: "修改", 3: "删除", 4: "导出", 5: "导入", 6: "其他",
	}

	for i, log := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), log.Title)
		f.SetCellValue(sheet, cellName(2, row), log.OperSummary)
		operTypeText := operTypeMap[log.OperType]
		if operTypeText == "" {
			operTypeText = "其他"
		}
		f.SetCellValue(sheet, cellName(3, row), operTypeText)
		f.SetCellValue(sheet, cellName(4, row), log.OperName)
		f.SetCellValue(sheet, cellName(5, row), log.RequestMethod)
		f.SetCellValue(sheet, cellName(6, row), log.OperUrl)
		f.SetCellValue(sheet, cellName(7, row), log.OperIp)
		f.SetCellValue(sheet, cellName(8, row), log.OperParam)
		f.SetCellValue(sheet, cellName(9, row), log.JsonResult)
		statusText := "成功"
		if log.Status == 1 {
			statusText = "失败"
		}
		f.SetCellValue(sheet, cellName(10, row), statusText)
		f.SetCellValue(sheet, cellName(11, row), log.ErrorMsg)
		f.SetCellValue(sheet, cellName(12, row), log.CostTime)
		if log.OperTime != nil {
			f.SetCellValue(sheet, cellName(13, row), log.OperTime.String())
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
