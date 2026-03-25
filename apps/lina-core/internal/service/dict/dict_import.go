package dict

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// TypeImportResult defines the result of import operation.
type TypeImportResult struct {
	Success  int                    // Number of successful imports
	Fail     int                    // Number of failed imports
	FailList []TypeImportFailItem   // Failure list
}

// TypeImportFailItem defines a single import failure.
type TypeImportFailItem struct {
	Row    int    // Row number
	Reason string // Failure reason
}

// TypeImport reads an Excel file and creates dict types from it.
// If updateSupport is true, existing records will be updated; otherwise, they will be skipped.
func (s *Service) TypeImport(ctx context.Context, fileReader io.Reader, updateSupport bool) (*TypeImportResult, error) {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return nil, gerror.New("无法解析 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, gerror.New("无法读取 Sheet1")
	}

	if len(rows) < 2 {
		return &TypeImportResult{}, nil
	}

	result := &TypeImportResult{}
	cols := dao.SysDictType.Columns()

	for i, row := range rows[1:] { // Skip header
		rowNum := i + 2
		if len(row) < 2 {
			result.Fail++
			result.FailList = append(result.FailList, TypeImportFailItem{
				Row:    rowNum,
				Reason: "字典名称和字典类型为必填项",
			})
			continue
		}

		name := row[0]
		typeVal := row[1]
		if name == "" || typeVal == "" {
			result.Fail++
			result.FailList = append(result.FailList, TypeImportFailItem{
				Row:    rowNum,
				Reason: "字典名称和字典类型不能为空",
			})
			continue
		}

		// Check if type exists
		var existing *entity.SysDictType
		err := dao.SysDictType.Ctx(ctx).
			Where(do.SysDictType{Type: typeVal}).
			WhereNull(cols.DeletedAt).
			Scan(&existing)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, TypeImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}

		// Parse status
		status := 1
		if len(row) > 2 {
			switch row[2] {
			case "停用", "0":
				status = 0
			}
		}

		// Parse remark
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		if existing != nil {
			// Type exists
			if !updateSupport {
				// Ignore mode: skip this record
				result.Fail++
				result.FailList = append(result.FailList, TypeImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("字典类型 '%s' 已存在", typeVal),
				})
				continue
			}
			// Overwrite mode: update existing record
			_, err = dao.SysDictType.Ctx(ctx).
				Where(do.SysDictType{Id: existing.Id}).
				Data(do.SysDictType{
					Name:      name,
					Status:    status,
					Remark:    remark,
					UpdatedAt: gtime.Now(),
				}).
				Update()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, TypeImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("更新失败: %v", err),
				})
				continue
			}
		} else {
			// Create new record
			_, err = dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
				Name:      name,
				Type:      typeVal,
				Status:    status,
				Remark:    remark,
				CreatedAt: gtime.Now(),
				UpdatedAt: gtime.Now(),
			}).Insert()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, TypeImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("插入失败: %v", err),
				})
				continue
			}
		}

		result.Success++
	}

	return result, nil
}

// GenerateTypeImportTemplate creates an Excel template for dict type import.
func (s *Service) GenerateTypeImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"字典名称", "字典类型", "状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Example row
	f.SetCellValue(sheet, cellName(1, 2), "用户性别")
	f.SetCellValue(sheet, cellName(2, 2), "sys_user_sex")
	f.SetCellValue(sheet, cellName(3, 2), "正常")
	f.SetCellValue(sheet, cellName(4, 2), "用户性别字典")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DataImportResult defines the result of dict data import operation.
type DataImportResult struct {
	Success  int                   // Number of successful imports
	Fail     int                   // Number of failed imports
	FailList []DataImportFailItem  // Failure list
}

// DataImportFailItem defines a single import failure.
type DataImportFailItem struct {
	Row    int    // Row number
	Reason string // Failure reason
}

// DataImport reads an Excel file and creates dict data from it.
// If updateSupport is true, existing records (matched by dictType+value) will be updated; otherwise, they will be skipped.
func (s *Service) DataImport(ctx context.Context, fileReader io.Reader, updateSupport bool) (*DataImportResult, error) {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return nil, gerror.New("无法解析 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, gerror.New("无法读取 Sheet1")
	}

	if len(rows) < 2 {
		return &DataImportResult{}, nil
	}

	result := &DataImportResult{}
	cols := dao.SysDictData.Columns()
	typeCols := dao.SysDictType.Columns()

	for i, row := range rows[1:] { // Skip header
		rowNum := i + 2
		if len(row) < 3 {
			result.Fail++
			result.FailList = append(result.FailList, DataImportFailItem{
				Row:    rowNum,
				Reason: "字典类型、字典标签、字典键值为必填项",
			})
			continue
		}

		dictType := row[0]
		label := row[1]
		value := row[2]
		if dictType == "" || label == "" || value == "" {
			result.Fail++
			result.FailList = append(result.FailList, DataImportFailItem{
				Row:    rowNum,
				Reason: "字典类型、字典标签、字典键值不能为空",
			})
			continue
		}

		// Verify dict type exists
		typeCount, err := dao.SysDictType.Ctx(ctx).
			Where(do.SysDictType{Type: dictType}).
			WhereNull(typeCols.DeletedAt).
			Count()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, DataImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}
		if typeCount == 0 {
			result.Fail++
			result.FailList = append(result.FailList, DataImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("字典类型 '%s' 不存在", dictType),
			})
			continue
		}

		// Check if dict data exists (by dictType + value)
		var existing *entity.SysDictData
		err = dao.SysDictData.Ctx(ctx).
			Where(do.SysDictData{DictType: dictType, Value: value}).
			WhereNull(cols.DeletedAt).
			Scan(&existing)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, DataImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}

		// Parse sort
		sort := 0
		if len(row) > 3 && row[3] != "" {
			fmt.Sscanf(row[3], "%d", &sort)
		}

		// Parse tagStyle
		tagStyle := ""
		if len(row) > 4 {
			tagStyle = row[4]
		}

		// Parse cssClass
		cssClass := ""
		if len(row) > 5 {
			cssClass = row[5]
		}

		// Parse status
		status := 1
		if len(row) > 6 {
			switch row[6] {
			case "停用", "0":
				status = 0
			}
		}

		// Parse remark
		remark := ""
		if len(row) > 7 {
			remark = row[7]
		}

		if existing != nil {
			// Record exists
			if !updateSupport {
				// Ignore mode: skip this record
				result.Fail++
				result.FailList = append(result.FailList, DataImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("字典数据 '%s/%s' 已存在", dictType, value),
				})
				continue
			}
			// Overwrite mode: update existing record
			_, err = dao.SysDictData.Ctx(ctx).
				Where(do.SysDictData{Id: existing.Id}).
				Data(do.SysDictData{
					Label:     label,
					Sort:      sort,
					TagStyle:  tagStyle,
					CssClass:  cssClass,
					Status:    status,
					Remark:    remark,
					UpdatedAt: gtime.Now(),
				}).
				Update()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, DataImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("更新失败: %v", err),
				})
				continue
			}
		} else {
			// Create new record
			_, err = dao.SysDictData.Ctx(ctx).Data(do.SysDictData{
				DictType:  dictType,
				Label:     label,
				Value:     value,
				Sort:      sort,
				TagStyle:  tagStyle,
				CssClass:  cssClass,
				Status:    status,
				Remark:    remark,
				CreatedAt: gtime.Now(),
				UpdatedAt: gtime.Now(),
			}).Insert()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, DataImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("插入失败: %v", err),
				})
				continue
			}
		}

		result.Success++
	}

	return result, nil
}

// GenerateDataImportTemplate creates an Excel template for dict data import.
func (s *Service) GenerateDataImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"字典类型", "字典标签", "字典键值", "排序", "Tag样式", "CSS类名", "状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Example row
	f.SetCellValue(sheet, cellName(1, 2), "sys_user_sex")
	f.SetCellValue(sheet, cellName(2, 2), "男")
	f.SetCellValue(sheet, cellName(3, 2), "1")
	f.SetCellValue(sheet, cellName(4, 2), "1")
	f.SetCellValue(sheet, cellName(5, 2), "primary")
	f.SetCellValue(sheet, cellName(6, 2), "")
	f.SetCellValue(sheet, cellName(7, 2), "正常")
	f.SetCellValue(sheet, cellName(8, 2), "男性")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}