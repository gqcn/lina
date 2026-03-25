package dict

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

// Service provides dict management operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum  int    // Page number, starting from 1
	PageSize int    // Page size
	Name     string // Dictionary name, supports fuzzy search
	Type     string // Dictionary type, supports fuzzy search
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysDictType // Dictionary type list
	Total int                   // Total count
}

// List queries dict type list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysDictType.Columns()
		m    = dao.SysDictType.Ctx(ctx).WhereNull(cols.DeletedAt)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Type != "" {
		m = m.WhereLike(cols.Type, "%"+in.Type+"%")
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysDictType
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

// CreateInput defines input for Create function.
type CreateInput struct {
	Name   string // Dictionary name
	Type   string // Dictionary type
	Status int    // Status: 1=Normal 0=Disabled
	Remark string // Remark
}

// Create creates a new dict type.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check type uniqueness
	cols := dao.SysDictType.Columns()
	count, err := dao.SysDictType.Ctx(ctx).
		Where(do.SysDictType{Type: in.Type}).
		WhereNull(cols.DeletedAt).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("字典类型已存在")
	}

	// Insert dict type
	id, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
		Name:      in.Name,
		Type:      in.Type,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetById retrieves dict type by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysDictType, error) {
	var dictType *entity.SysDictType
	cols := dao.SysDictType.Columns()
	err := dao.SysDictType.Ctx(ctx).
		Where(do.SysDictType{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&dictType)
	if err != nil {
		return nil, err
	}
	if dictType == nil {
		return nil, gerror.New("字典类型不存在")
	}
	return dictType, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int      // Dictionary type ID
	Name   *string  // Dictionary name
	Type   *string  // Dictionary type
	Status *int     // Status: 1=Normal 0=Disabled
	Remark *string  // Remark
}

// Update updates dict type information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check dict type exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysDictType{
		UpdatedAt: gtime.Now(),
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Type != nil {
		// Check type uniqueness when updating the type field
		if *in.Type != "" {
			cols := dao.SysDictType.Columns()
			count, err := dao.SysDictType.Ctx(ctx).
				Where(cols.Type, *in.Type).
				WhereNull(cols.DeletedAt).
				WhereNot(cols.Id, in.Id).
				Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return gerror.New("字典类型已存在")
			}
		}
		data.Type = *in.Type
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysDictType.Ctx(ctx).Where(do.SysDictType{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a dict type.
func (s *Service) Delete(ctx context.Context, id int) error {
	// Check dict type exists
	dictType, err := s.GetById(ctx, id)
	if err != nil {
		return err
	}

	// Check if dict_data exists for this type
	dataCols := dao.SysDictData.Columns()
	count, err := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{DictType: dictType.Type}).
		WhereNull(dataCols.DeletedAt).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("该字典类型下存在字典数据，不能删除")
	}

	// Soft delete
	_, err = dao.SysDictType.Ctx(ctx).
		Where(do.SysDictType{Id: id}).
		Data(do.SysDictType{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Name string // Dictionary name, supports fuzzy search
	Type string // Dictionary type, supports fuzzy search
	Ids  []int  // Specific IDs to export; if empty, export all matching records
}

// Export generates an Excel file with dict type data (max 10000 rows).
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysDictType.Columns()
	m := dao.SysDictType.Ctx(ctx).WhereNull(cols.DeletedAt)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.Name != "" {
			m = m.WhereLike(cols.Name, "%"+in.Name+"%")
		}
		if in.Type != "" {
			m = m.WhereLike(cols.Type, "%"+in.Type+"%")
		}
	}

	// Limit export to prevent memory issues
	m = m.Limit(10000)

	var list []*entity.SysDictType
	err := m.Order(cols.Id + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"字典名称", "字典类型", "状态", "备注", "创建时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, dt := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), dt.Name)
		f.SetCellValue(sheet, cellName(2, row), dt.Type)
		statusText := "正常"
		if dt.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(sheet, cellName(3, row), statusText)
		f.SetCellValue(sheet, cellName(4, row), dt.Remark)
		if dt.CreatedAt != nil {
			f.SetCellValue(sheet, cellName(5, row), dt.CreatedAt.String())
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// OptionItem defines a single option item.
type OptionItem struct {
	Id   int    `json:"id"`   // Dictionary type ID
	Name string `json:"name"` // Dictionary name
	Type string `json:"type"` // Dictionary type
}

// Options returns all non-deleted dict types with status=1.
func (s *Service) Options(ctx context.Context) ([]*OptionItem, error) {
	cols := dao.SysDictType.Columns()
	var list []*entity.SysDictType
	err := dao.SysDictType.Ctx(ctx).
		WhereNull(cols.DeletedAt).
		Where(do.SysDictType{Status: 1}).
		Order(cols.Id + " ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	options := make([]*OptionItem, 0, len(list))
	for _, dt := range list {
		options = append(options, &OptionItem{
			Id:   dt.Id,
			Name: dt.Name,
			Type: dt.Type,
		})
	}
	return options, nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
