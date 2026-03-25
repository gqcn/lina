package sysconfig

import (
	"bytes"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Service provides system config management operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum   int    // Page number, starting from 1
	PageSize  int    // Page size
	Name      string // Parameter name, supports fuzzy search
	Key       string // Parameter key, supports fuzzy search
	BeginTime string // Creation time start
	EndTime   string // Creation time end
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysConfig // Config list
	Total int                 // Total count
}

// List queries config list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysConfig.Columns()
		m    = dao.SysConfig.Ctx(ctx).WhereNull(cols.DeletedAt)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Key != "" {
		m = m.WhereLike(cols.Key, "%"+in.Key+"%")
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime+" 00:00:00")
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime+" 23:59:59")
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysConfig
	err = m.Page(in.PageNum, in.PageSize).
		Order(cols.Id + " DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves config by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	cols := dao.SysConfig.Columns()
	err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, gerror.New("参数设置不存在")
	}
	return cfg, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name   string // Parameter name
	Key    string // Parameter key
	Value  string // Parameter value
	Remark string // Remark
}

// Create creates a new config record.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check key uniqueness
	cols := dao.SysConfig.Columns()
	count, err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Key: in.Key}).
		WhereNull(cols.DeletedAt).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("参数键名已存在")
	}

	// Insert config
	id, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		Name:      in.Name,
		Key:       in.Key,
		Value:     in.Value,
		Remark:    in.Remark,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int      // Parameter ID
	Name   *string  // Parameter name
	Key    *string  // Parameter key
	Value  *string  // Parameter value
	Remark *string  // Remark
}

// Update updates config information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check config exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	// Check key uniqueness (exclude self)
	if in.Key != nil {
		cols := dao.SysConfig.Columns()
		count, err := dao.SysConfig.Ctx(ctx).
			Where(do.SysConfig{Key: *in.Key}).
			WhereNot(cols.Id, in.Id).
			WhereNull(cols.DeletedAt).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.New("参数键名已存在")
		}
	}

	data := do.SysConfig{
		UpdatedAt: gtime.Now(),
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Key != nil {
		data.Key = *in.Key
	}
	if in.Value != nil {
		data.Value = *in.Value
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a config record.
func (s *Service) Delete(ctx context.Context, id int) error {
	// Check config exists
	if _, err := s.GetById(ctx, id); err != nil {
		return err
	}

	// Soft delete
	_, err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: id}).
		Data(do.SysConfig{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// GetByKey retrieves config by key name.
func (s *Service) GetByKey(ctx context.Context, key string) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	cols := dao.SysConfig.Columns()
	err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Key: key}).
		WhereNull(cols.DeletedAt).
		Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, gerror.New("参数键名不存在")
	}
	return cfg, nil
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Name      string // Parameter name, supports fuzzy search
	Key       string // Parameter key, supports fuzzy search
	BeginTime string // Creation time start
	EndTime   string // Creation time end
	Ids       []int  // Specific IDs to export; if empty, export all matching records
}

// Export generates an Excel file with config data.
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysConfig.Columns()
	m := dao.SysConfig.Ctx(ctx).WhereNull(cols.DeletedAt)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.Name != "" {
			m = m.WhereLike(cols.Name, "%"+in.Name+"%")
		}
		if in.Key != "" {
			m = m.WhereLike(cols.Key, "%"+in.Key+"%")
		}
		if in.BeginTime != "" {
			m = m.WhereGTE(cols.CreatedAt, in.BeginTime+" 00:00:00")
		}
		if in.EndTime != "" {
			m = m.WhereLTE(cols.CreatedAt, in.EndTime+" 23:59:59")
		}
	}

	var list []*entity.SysConfig
	err := m.Order(cols.Id + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"参数名称", "参数键名", "参数键值", "备注", "创建时间", "修改时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, c := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), c.Name)
		f.SetCellValue(sheet, cellName(2, row), c.Key)
		f.SetCellValue(sheet, cellName(3, row), c.Value)
		f.SetCellValue(sheet, cellName(4, row), c.Remark)
		if c.CreatedAt != nil {
			f.SetCellValue(sheet, cellName(5, row), c.CreatedAt.String())
		}
		if c.UpdatedAt != nil {
			f.SetCellValue(sheet, cellName(6, row), c.UpdatedAt.String())
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
