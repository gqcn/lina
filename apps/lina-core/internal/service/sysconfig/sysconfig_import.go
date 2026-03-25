package sysconfig

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

// ImportResult defines the result of config import operation.
type ImportResult struct {
	Success  int                 // Number of successful imports
	Fail     int                 // Number of failed imports
	FailList []ImportFailItem    // Failure list
}

// ImportFailItem defines a single import failure.
type ImportFailItem struct {
	Row    int    // Row number
	Reason string // Failure reason
}

// Import reads an Excel file and creates configs from it.
// If updateSupport is true, existing records (matched by key) will be updated; otherwise, they will be skipped.
func (s *Service) Import(ctx context.Context, fileReader io.Reader, updateSupport bool) (*ImportResult, error) {
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
		return &ImportResult{}, nil
	}

	result := &ImportResult{}
	cols := dao.SysConfig.Columns()

	for i, row := range rows[1:] { // Skip header
		rowNum := i + 2
		if len(row) < 3 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "参数名称、参数键名、参数键值为必填项",
			})
			continue
		}

		name := row[0]
		key := row[1]
		value := row[2]
		if name == "" || key == "" || value == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "参数名称、参数键名、参数键值不能为空",
			})
			continue
		}

		// Check if key exists
		var existing *entity.SysConfig
		err := dao.SysConfig.Ctx(ctx).
			Where(do.SysConfig{Key: key}).
			WhereNull(cols.DeletedAt).
			Scan(&existing)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}

		// Parse remark
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		if existing != nil {
			// Key exists
			if !updateSupport {
				// Ignore mode: skip this record
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("参数键名 '%s' 已存在", key),
				})
				continue
			}
			// Overwrite mode: update existing record
			_, err = dao.SysConfig.Ctx(ctx).
				Where(do.SysConfig{Id: existing.Id}).
				Data(do.SysConfig{
					Name:      name,
					Value:     value,
					Remark:    remark,
					UpdatedAt: gtime.Now(),
				}).
				Update()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
					Row:    rowNum,
					Reason: fmt.Sprintf("更新失败: %v", err),
				})
				continue
			}
		} else {
			// Create new record
			_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
				Name:      name,
				Key:       key,
				Value:     value,
				Remark:    remark,
				CreatedAt: gtime.Now(),
				UpdatedAt: gtime.Now(),
			}).Insert()
			if err != nil {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItem{
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

// GenerateImportTemplate creates an Excel template for config import.
func (s *Service) GenerateImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"参数名称", "参数键名", "参数键值", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Example row
	f.SetCellValue(sheet, cellName(1, 2), "系统名称")
	f.SetCellValue(sheet, cellName(2, 2), "sys.app.name")
	f.SetCellValue(sheet, cellName(3, 2), "Lina")
	f.SetCellValue(sheet, cellName(4, 2), "系统显示名称")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}