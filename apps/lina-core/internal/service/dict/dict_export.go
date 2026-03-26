package dict

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

// CombinedExportInput defines input for CombinedExport function.
type CombinedExportInput struct {
	Name string // Dictionary name, supports fuzzy search
	Type string // Dictionary type, supports fuzzy search
	Ids  []int  // Specific type IDs to export; if empty, export all matching types
}

// CombinedExport generates an Excel file with both dict types and dict data (max 10000 rows each).
func (s *Service) CombinedExport(ctx context.Context, in CombinedExportInput) ([]byte, error) {
	// Query dict types
	typeCols := dao.SysDictType.Columns()
	typeM := dao.SysDictType.Ctx(ctx)

	if len(in.Ids) > 0 {
		typeM = typeM.WhereIn(typeCols.Id, in.Ids)
	} else {
		if in.Name != "" {
			typeM = typeM.WhereLike(typeCols.Name, "%"+in.Name+"%")
		}
		if in.Type != "" {
			typeM = typeM.WhereLike(typeCols.Type, "%"+in.Type+"%")
		}
	}

	typeM = typeM.Limit(10000)

	var typeList []*entity.SysDictType
	err := typeM.Order(typeCols.Id + " ASC").Scan(&typeList)
	if err != nil {
		return nil, err
	}

	// Collect dict type strings for querying dict data
	typeStrings := make([]string, 0, len(typeList))
	for _, t := range typeList {
		typeStrings = append(typeStrings, t.Type)
	}

	// Query dict data for the selected types
	var dataList []*entity.SysDictData
	if len(typeStrings) > 0 {
		dataCols := dao.SysDictData.Columns()
		dataM := dao.SysDictData.Ctx(ctx).
			WhereIn(dataCols.DictType, typeStrings).
			Limit(10000)

		err = dataM.Order(dataCols.Sort + " ASC").Scan(&dataList)
		if err != nil {
			return nil, err
		}
	}

	// Create Excel file with two sheets
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: 字典类型
	typeSheet := "字典类型"
	f.SetSheetName("Sheet1", typeSheet)

	typeHeaders := []string{"字典名称", "字典类型", "状态", "备注", "创建时间"}
	for i, h := range typeHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(typeSheet, cell, h)
	}

	for i, dt := range typeList {
		row := i + 2
		f.SetCellValue(typeSheet, cellName(1, row), dt.Name)
		f.SetCellValue(typeSheet, cellName(2, row), dt.Type)
		statusText := "正常"
		if dt.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(typeSheet, cellName(3, row), statusText)
		f.SetCellValue(typeSheet, cellName(4, row), dt.Remark)
		if dt.CreatedAt != nil {
			f.SetCellValue(typeSheet, cellName(5, row), dt.CreatedAt.String())
		}
	}

	// Sheet 2: 字典数据
	dataSheet := "字典数据"
	f.NewSheet(dataSheet)

	dataHeaders := []string{"所属类型", "字典标签", "字典值", "排序", "Tag样式", "CSS类", "状态", "备注", "创建时间"}
	for i, h := range dataHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(dataSheet, cell, h)
	}

	for i, dd := range dataList {
		row := i + 2
		f.SetCellValue(dataSheet, cellName(1, row), dd.DictType)
		f.SetCellValue(dataSheet, cellName(2, row), dd.Label)
		f.SetCellValue(dataSheet, cellName(3, row), dd.Value)
		f.SetCellValue(dataSheet, cellName(4, row), dd.Sort)
		f.SetCellValue(dataSheet, cellName(5, row), dd.TagStyle)
		f.SetCellValue(dataSheet, cellName(6, row), dd.CssClass)
		statusText := "正常"
		if dd.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(dataSheet, cellName(7, row), statusText)
		f.SetCellValue(dataSheet, cellName(8, row), dd.Remark)
		if dd.CreatedAt != nil {
			f.SetCellValue(dataSheet, cellName(9, row), dd.CreatedAt.String())
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}