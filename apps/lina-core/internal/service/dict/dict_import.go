package dict

import (
	"bytes"
	"context"
	"io"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// CombinedImportResult represents the result of combined import.
type CombinedImportResult struct {
	TypeSuccess int
	TypeFail    int
	DataSuccess int
	DataFail    int
	FailList    []ImportFailItem
}

// ImportFailItem represents a failed import record.
type ImportFailItem struct {
	Sheet  string
	Row    int
	Reason string
}

// CombinedImport imports dictionary types and data from an Excel file.
// If updateSupport is true, existing records will be updated; otherwise, they will be skipped.
func (s *Service) CombinedImport(ctx context.Context, fileData []byte, updateSupport bool) (*CombinedImportResult, error) {
	result := &CombinedImportResult{
		FailList: make([]ImportFailItem, 0),
	}

	// Open Excel file
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, gerror.New("无法解析Excel文件")
	}
	defer f.Close()

	// Get existing dict types for validation
	typeCols := dao.SysDictType.Columns()
	existingTypes := make(map[string]bool)
	var existingTypeList []*struct {
		Type string
	}
	err = dao.SysDictType.Ctx(ctx).
		WhereNull(typeCols.DeletedAt).
		Fields(typeCols.Type).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Import Sheet 1: 字典类型
	typeSheet := "字典类型"
	typeRows, err := f.GetRows(typeSheet)
	if err != nil {
		// Sheet might not exist, skip
		typeRows = nil
	}

	// Track imported types for data import
	importedTypes := make(map[string]bool)

	for i, row := range typeRows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 3 { // Need at least: 名称, 类型, 状态
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: "数据不完整",
			})
			continue
		}

		name := row[0]
		typeStr := row[1]
		status := 1
		if len(row) > 2 && row[2] == "停用" {
			status = 0
		}
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		// Check if type already exists
		if existingTypes[typeStr] {
			if updateSupport {
				// Update existing record
				_, err := dao.SysDictType.Ctx(ctx).
					Where(do.SysDictType{Type: typeStr}).
					Data(do.SysDictType{
						Name:      name,
						Status:    status,
						Remark:    remark,
						UpdatedAt: gtime.Now(),
					}).Update()
				if err != nil {
					result.TypeFail++
					result.FailList = append(result.FailList, ImportFailItem{
						Sheet:  typeSheet,
						Row:    i + 1,
						Reason: "更新失败: " + err.Error(),
					})
					continue
				}
				importedTypes[typeStr] = true
				result.TypeSuccess++
			} else {
				result.TypeFail++
				result.FailList = append(result.FailList, ImportFailItem{
					Sheet:  typeSheet,
					Row:    i + 1,
					Reason: "字典类型已存在",
				})
			}
			continue
		}

		// Insert dict type
		_, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
			Name:      name,
			Type:      typeStr,
			Status:    status,
			Remark:    remark,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}).InsertAndGetId()
		if err != nil {
			result.TypeFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  typeSheet,
				Row:    i + 1,
				Reason: "插入失败: " + err.Error(),
			})
			continue
		}

		existingTypes[typeStr] = true
		importedTypes[typeStr] = true
		result.TypeSuccess++
	}

	// Import Sheet 2: 字典数据
	dataSheet := "字典数据"
	dataRows, err := f.GetRows(dataSheet)
	if err != nil {
		// Sheet might not exist, skip
		dataRows = nil
	}

	for i, row := range dataRows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 4 { // Need at least: 所属类型, 标签, 值, 排序
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: "数据不完整",
			})
			continue
		}

		dictType := row[0]
		label := row[1]
		value := row[2]
		sort := 0
		if len(row) > 3 && row[3] != "" {
			// Parse sort
			for _, c := range row[3] {
				if c >= '0' && c <= '9' {
					sort = sort*10 + int(c-'0')
				}
			}
		}
		tagStyle := ""
		if len(row) > 4 {
			tagStyle = row[4]
		}
		cssClass := ""
		if len(row) > 5 {
			cssClass = row[5]
		}
		status := 1
		if len(row) > 6 && row[6] == "停用" {
			status = 0
		}
		remark := ""
		if len(row) > 7 {
			remark = row[7]
		}

		// Check if dict_type exists
		if !existingTypes[dictType] {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: "字典类型不存在",
			})
			continue
		}

		// Check if dict_data already exists (dict_type + value unique)
		dataCols := dao.SysDictData.Columns()
		var existingData *entity.SysDictData
		err = dao.SysDictData.Ctx(ctx).
			Where(do.SysDictData{DictType: dictType, Value: value}).
			WhereNull(dataCols.DeletedAt).
			Scan(&existingData)
		if err != nil {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: "查询失败: " + err.Error(),
			})
			continue
		}

		if existingData != nil {
			if updateSupport {
				// Update existing record
				_, err := dao.SysDictData.Ctx(ctx).
					Where(do.SysDictData{Id: existingData.Id}).
					Data(do.SysDictData{
						Label:     label,
						Sort:      sort,
						TagStyle:  tagStyle,
						CssClass:  cssClass,
						Status:    status,
						Remark:    remark,
						UpdatedAt: gtime.Now(),
					}).Update()
				if err != nil {
					result.DataFail++
					result.FailList = append(result.FailList, ImportFailItem{
						Sheet:  dataSheet,
						Row:    i + 1,
						Reason: "更新失败: " + err.Error(),
					})
					continue
				}
				result.DataSuccess++
			} else {
				result.DataFail++
				result.FailList = append(result.FailList, ImportFailItem{
					Sheet:  dataSheet,
					Row:    i + 1,
					Reason: "字典值已存在",
				})
			}
			continue
		}

		// Insert dict data
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
		}).InsertAndGetId()
		if err != nil {
			result.DataFail++
			result.FailList = append(result.FailList, ImportFailItem{
				Sheet:  dataSheet,
				Row:    i + 1,
				Reason: "插入失败: " + err.Error(),
			})
			continue
		}

		result.DataSuccess++
	}

	return result, nil
}

// CombinedImportTemplate generates an Excel template for dictionary import.
func (s *Service) CombinedImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: 字典类型
	typeSheet := "字典类型"
	f.SetSheetName("Sheet1", typeSheet)

	typeHeaders := []string{"字典名称", "字典类型", "状态", "备注"}
	for i, h := range typeHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(typeSheet, cell, h)
	}

	// Add example row
	f.SetCellValue(typeSheet, "A2", "用户性别")
	f.SetCellValue(typeSheet, "B2", "sys_user_sex")
	f.SetCellValue(typeSheet, "C2", "正常")
	f.SetCellValue(typeSheet, "D2", "用户性别字典")

	// Sheet 2: 字典数据
	dataSheet := "字典数据"
	f.NewSheet(dataSheet)

	dataHeaders := []string{"所属类型", "字典标签", "字典值", "排序", "Tag样式", "CSS类", "状态", "备注"}
	for i, h := range dataHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(dataSheet, cell, h)
	}

	// Add example rows
	f.SetCellValue(dataSheet, "A2", "sys_user_sex")
	f.SetCellValue(dataSheet, "B2", "男")
	f.SetCellValue(dataSheet, "C2", "1")
	f.SetCellValue(dataSheet, "D2", "1")
	f.SetCellValue(dataSheet, "E2", "primary")
	f.SetCellValue(dataSheet, "F2", "")
	f.SetCellValue(dataSheet, "G2", "正常")
	f.SetCellValue(dataSheet, "H2", "男性")

	f.SetCellValue(dataSheet, "A3", "sys_user_sex")
	f.SetCellValue(dataSheet, "B3", "女")
	f.SetCellValue(dataSheet, "C3", "2")
	f.SetCellValue(dataSheet, "D3", "2")
	f.SetCellValue(dataSheet, "E3", "danger")
	f.SetCellValue(dataSheet, "F3", "")
	f.SetCellValue(dataSheet, "G3", "正常")
	f.SetCellValue(dataSheet, "H3", "女性")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ImportResult represents the result of import operation.
type ImportResult struct {
	Success  int
	Fail     int
	FailList []ImportFailItemRecord
}

// ImportFailItemRecord represents a failed import record.
type ImportFailItemRecord struct {
	Row    int
	Reason string
}

// TypeImport imports dictionary types from an Excel file.
func (s *Service) TypeImport(ctx context.Context, file io.Reader, updateSupport bool) (*ImportResult, error) {
	result := &ImportResult{
		FailList: make([]ImportFailItemRecord, 0),
	}

	// Open Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, gerror.New("无法解析Excel文件")
	}
	defer f.Close()

	// Get existing dict types
	typeCols := dao.SysDictType.Columns()
	existingTypes := make(map[string]bool)
	var existingTypeList []*entity.SysDictType
	err = dao.SysDictType.Ctx(ctx).
		WhereNull(typeCols.DeletedAt).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Read Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, gerror.New("无法读取Excel文件")
	}

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 2 { // Need at least: 名称, 类型
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "数据不完整",
			})
			continue
		}

		name := row[0]
		typeStr := row[1]
		status := 1
		if len(row) > 2 && row[2] == "停用" {
			status = 0
		}
		remark := ""
		if len(row) > 3 {
			remark = row[3]
		}

		// Check if type already exists
		if existingTypes[typeStr] {
			if updateSupport {
				// Update existing record
				_, err := dao.SysDictType.Ctx(ctx).
					Where(do.SysDictType{Type: typeStr}).
					Data(do.SysDictType{
						Name:      name,
						Status:    status,
						Remark:    remark,
						UpdatedAt: gtime.Now(),
					}).Update()
				if err != nil {
					result.Fail++
					result.FailList = append(result.FailList, ImportFailItemRecord{
						Row:    i + 1,
						Reason: "更新失败: " + err.Error(),
					})
					continue
				}
				result.Success++
			} else {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItemRecord{
					Row:    i + 1,
					Reason: "字典类型已存在",
				})
			}
			continue
		}

		// Insert new record
		_, err := dao.SysDictType.Ctx(ctx).Data(do.SysDictType{
			Name:      name,
			Type:      typeStr,
			Status:    status,
			Remark:    remark,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}).InsertAndGetId()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "插入失败: " + err.Error(),
			})
			continue
		}

		existingTypes[typeStr] = true
		result.Success++
	}

	return result, nil
}

// DataImport imports dictionary data from an Excel file.
func (s *Service) DataImport(ctx context.Context, file io.Reader, updateSupport bool) (*ImportResult, error) {
	result := &ImportResult{
		FailList: make([]ImportFailItemRecord, 0),
	}

	// Open Excel file
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, gerror.New("无法解析Excel文件")
	}
	defer f.Close()

	// Get existing dict types
	typeCols := dao.SysDictType.Columns()
	existingTypes := make(map[string]bool)
	var existingTypeList []*entity.SysDictType
	err = dao.SysDictType.Ctx(ctx).
		WhereNull(typeCols.DeletedAt).
		Scan(&existingTypeList)
	if err != nil {
		return nil, err
	}
	for _, t := range existingTypeList {
		existingTypes[t.Type] = true
	}

	// Read Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, gerror.New("无法读取Excel文件")
	}

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}
		if len(row) < 4 { // Need at least: 所属类型, 标签, 值, 排序
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "数据不完整",
			})
			continue
		}

		dictType := row[0]
		label := row[1]
		value := row[2]
		sort := 0
		if len(row) > 3 && row[3] != "" {
			for _, c := range row[3] {
				if c >= '0' && c <= '9' {
					sort = sort*10 + int(c-'0')
				}
			}
		}
		tagStyle := ""
		if len(row) > 4 {
			tagStyle = row[4]
		}
		cssClass := ""
		if len(row) > 5 {
			cssClass = row[5]
		}
		status := 1
		if len(row) > 6 && row[6] == "停用" {
			status = 0
		}
		remark := ""
		if len(row) > 7 {
			remark = row[7]
		}

		// Check if dict_type exists
		if !existingTypes[dictType] {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "字典类型不存在",
			})
			continue
		}

		// Check if dict_data already exists
		dataCols := dao.SysDictData.Columns()
		var existingData *entity.SysDictData
		err = dao.SysDictData.Ctx(ctx).
			Where(do.SysDictData{DictType: dictType, Value: value}).
			WhereNull(dataCols.DeletedAt).
			Scan(&existingData)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "查询失败: " + err.Error(),
			})
			continue
		}

		if existingData != nil {
			if updateSupport {
				// Update existing record
				_, err := dao.SysDictData.Ctx(ctx).
					Where(do.SysDictData{Id: existingData.Id}).
					Data(do.SysDictData{
						Label:     label,
						Sort:      sort,
						TagStyle:  tagStyle,
						CssClass:  cssClass,
						Status:    status,
						Remark:    remark,
						UpdatedAt: gtime.Now(),
					}).Update()
				if err != nil {
					result.Fail++
					result.FailList = append(result.FailList, ImportFailItemRecord{
						Row:    i + 1,
						Reason: "更新失败: " + err.Error(),
					})
					continue
				}
				result.Success++
			} else {
				result.Fail++
				result.FailList = append(result.FailList, ImportFailItemRecord{
					Row:    i + 1,
					Reason: "字典值已存在",
				})
			}
			continue
		}

		// Insert new record
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
		}).InsertAndGetId()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItemRecord{
				Row:    i + 1,
				Reason: "插入失败: " + err.Error(),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateTypeImportTemplate generates an Excel template for dictionary type import.
func (s *Service) GenerateTypeImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	headers := []string{"字典名称", "字典类型", "状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Add example row
	f.SetCellValue(sheet, "A2", "用户性别")
	f.SetCellValue(sheet, "B2", "sys_user_sex")
	f.SetCellValue(sheet, "C2", "正常")
	f.SetCellValue(sheet, "D2", "用户性别字典")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateDataImportTemplate generates an Excel template for dictionary data import.
func (s *Service) GenerateDataImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	headers := []string{"所属类型", "字典标签", "字典值", "排序", "Tag样式", "CSS类", "状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Add example rows
	f.SetCellValue(sheet, "A2", "sys_user_sex")
	f.SetCellValue(sheet, "B2", "男")
	f.SetCellValue(sheet, "C2", "1")
	f.SetCellValue(sheet, "D2", "1")
	f.SetCellValue(sheet, "E2", "primary")
	f.SetCellValue(sheet, "F2", "")
	f.SetCellValue(sheet, "G2", "正常")
	f.SetCellValue(sheet, "H2", "男性")

	f.SetCellValue(sheet, "A3", "sys_user_sex")
	f.SetCellValue(sheet, "B3", "女")
	f.SetCellValue(sheet, "C3", "2")
	f.SetCellValue(sheet, "D3", "2")
	f.SetCellValue(sheet, "E3", "danger")
	f.SetCellValue(sheet, "F3", "")
	f.SetCellValue(sheet, "G3", "正常")
	f.SetCellValue(sheet, "H3", "女性")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
