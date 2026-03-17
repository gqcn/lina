package post

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

// ExportInput defines input for Export function.
type ExportInput struct {
	DeptId *int
	Code   string
	Name   string
	Status *int
}

// Export generates an Excel file with post data based on filters.
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysPost.Columns()
	m := dao.SysPost.Ctx(ctx).WhereNull(cols.DeletedAt)

	// Apply filters
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			m = m.Where(cols.DeptId, 0)
		} else {
			deptIds, err := s.getDeptAndDescendantIds(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			m = m.WhereIn(cols.DeptId, deptIds)
		}
	}
	if in.Code != "" {
		m = m.WhereLike(cols.Code, "%"+in.Code+"%")
	}
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	var list []*entity.SysPost
	err := m.Order(cols.Sort + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"岗位编码", "岗位名称", "排序", "状态", "备注", "创建时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, p := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), p.Code)
		f.SetCellValue(sheet, cellName(2, row), p.Name)
		f.SetCellValue(sheet, cellName(3, row), p.Sort)
		statusText := "正常"
		if p.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(sheet, cellName(4, row), statusText)
		f.SetCellValue(sheet, cellName(5, row), p.Remark)
		if p.CreatedAt != nil {
			f.SetCellValue(sheet, cellName(6, row), p.CreatedAt.String())
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
