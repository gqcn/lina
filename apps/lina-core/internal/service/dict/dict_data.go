package dict

import (
	"bytes"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// DataListInput defines input for DataList function.
type DataListInput struct {
	PageNum  int
	PageSize int
	DictType string
	Label    string
}

// DataListOutput defines output for DataList function.
type DataListOutput struct {
	List  []*entity.SysDictData
	Total int
}

// DataList queries dict data list with pagination and filters.
func (s *Service) DataList(ctx context.Context, in DataListInput) (*DataListOutput, error) {
	var (
		cols = dao.SysDictData.Columns()
		m    = dao.SysDictData.Ctx(ctx)
	)

	// Apply filters
	if in.DictType != "" {
		m = m.Where(do.SysDictData{DictType: in.DictType})
	}
	if in.Label != "" {
		m = m.WhereLike(cols.Label, "%"+in.Label+"%")
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysDictData
	err = m.Page(in.PageNum, in.PageSize).
		Order(cols.Sort + " ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &DataListOutput{
		List:  list,
		Total: total,
	}, nil
}

// DataCreateInput defines input for DataCreate function.
type DataCreateInput struct {
	DictType string
	Label    string
	Value    string
	Sort     int
	TagStyle string
	CssClass string
	Status   int
	Remark   string
}

// DataCreate creates a new dict data entry.
func (s *Service) DataCreate(ctx context.Context, in DataCreateInput) (int, error) {
	id, err := dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
		DictType: in.DictType,
		Label:    in.Label,
		Value:    in.Value,
		Sort:     in.Sort,
		TagStyle: in.TagStyle,
		CssClass: in.CssClass,
		Status:   in.Status,
		Remark:   in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// DataGetById retrieves dict data by ID.
func (s *Service) DataGetById(ctx context.Context, id int) (*entity.SysDictData, error) {
	var dictData *entity.SysDictData
	err := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{Id: id}).
		Scan(&dictData)
	if err != nil {
		return nil, err
	}
	if dictData == nil {
		return nil, gerror.New("字典数据不存在")
	}
	return dictData, nil
}

// DataUpdateInput defines input for DataUpdate function.
type DataUpdateInput struct {
	Id       int
	DictType *string
	Label    *string
	Value    *string
	Sort     *int
	TagStyle *string
	CssClass *string
	Status   *int
	Remark   *string
}

// DataUpdate updates dict data information.
func (s *Service) DataUpdate(ctx context.Context, in DataUpdateInput) error {
	// Check dict data exists
	if _, err := s.DataGetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysDictData{}
	if in.DictType != nil {
		data.DictType = *in.DictType
	}
	if in.Label != nil {
		data.Label = *in.Label
	}
	if in.Value != nil {
		data.Value = *in.Value
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.TagStyle != nil {
		data.TagStyle = *in.TagStyle
	}
	if in.CssClass != nil {
		data.CssClass = *in.CssClass
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysDictData.Ctx(ctx).Where(do.SysDictData{Id: in.Id}).Data(data).Update()
	return err
}

// DataDelete hard-deletes a dict data entry.
func (s *Service) DataDelete(ctx context.Context, id int) error {
	// Check dict data exists
	if _, err := s.DataGetById(ctx, id); err != nil {
		return err
	}

	// Hard delete
	_, err := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{Id: id}).
		Delete()
	return err
}

// DataExportInput defines input for DataExport function.
type DataExportInput struct {
	DictType string
	Label    string
	Ids      []int // Specific IDs to export; if empty, export all matching records
}

// DataExport generates an Excel file with dict data (max 10000 rows).
func (s *Service) DataExport(ctx context.Context, in DataExportInput) ([]byte, error) {
	cols := dao.SysDictData.Columns()
	m := dao.SysDictData.Ctx(ctx)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.DictType != "" {
			m = m.Where(cols.DictType, in.DictType)
		}
		if in.Label != "" {
			m = m.WhereLike(cols.Label, "%"+in.Label+"%")
		}
	}

	// Limit export to prevent memory issues
	m = m.Limit(10000)

	var list []*entity.SysDictData
	err := m.Order(cols.Sort + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"字典标签", "字典值", "排序", "Tag样式", "CSS类", "状态", "备注", "创建时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, dd := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), dd.Label)
		f.SetCellValue(sheet, cellName(2, row), dd.Value)
		f.SetCellValue(sheet, cellName(3, row), dd.Sort)
		f.SetCellValue(sheet, cellName(4, row), dd.TagStyle)
		f.SetCellValue(sheet, cellName(5, row), dd.CssClass)
		statusText := "正常"
		if dd.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(sheet, cellName(6, row), statusText)
		f.SetCellValue(sheet, cellName(7, row), dd.Remark)
		if dd.CreatedAt != nil {
			f.SetCellValue(sheet, cellName(8, row), dd.CreatedAt.String())
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DataByType returns all non-deleted dict data for a given dict type with status=1, ordered by sort ASC.
func (s *Service) DataByType(ctx context.Context, dictType string) ([]*entity.SysDictData, error) {
	cols := dao.SysDictData.Columns()
	var list []*entity.SysDictData
	err := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{DictType: dictType, Status: 1}).
		Order(cols.Sort + " ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}
