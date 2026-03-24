package user

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
	"lina-core/internal/service/auth"
)

// ExportInput defines input for Export function.
type ExportInput struct {
	Ids []int // 用户ID列表，为空则导出全部
}

// Export generates an Excel file with user data based on IDs.
func (s *Service) Export(ctx context.Context, in ExportInput) ([]byte, error) {
	cols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).WhereNull(cols.DeletedAt)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	}

	var list []*entity.SysUser
	err := m.FieldsEx(cols.Password).
		Order(cols.Id + " ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"用户名", "昵称", "手机号码", "邮箱", "性别", "状态", "备注", "创建时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, u := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), u.Username)
		f.SetCellValue(sheet, cellName(2, row), u.Nickname)
		f.SetCellValue(sheet, cellName(3, row), u.Phone)
		f.SetCellValue(sheet, cellName(4, row), u.Email)
		sexText := "未知"
		switch u.Sex {
		case 1:
			sexText = "男"
		case 2:
			sexText = "女"
		}
		f.SetCellValue(sheet, cellName(5, row), sexText)
		statusText := "正常"
		if u.Status == 0 {
			statusText = "停用"
		}
		f.SetCellValue(sheet, cellName(6, row), statusText)
		f.SetCellValue(sheet, cellName(7, row), u.Remark)
		if u.CreatedAt != nil {
			f.SetCellValue(sheet, cellName(8, row), u.CreatedAt.String())
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ImportResult defines the result of import operation.
type ImportResult struct {
	Success  int              // 成功导入数量
	Fail     int              // 失败数量
	FailList []ImportFailItem // 失败列表
}

// ImportFailItem defines a single import failure.
type ImportFailItem struct {
	Row    int    // 行号
	Reason string // 失败原因
}

// Import reads an Excel file and creates users from it.
func (s *Service) Import(ctx context.Context, fileReader io.Reader) (*ImportResult, error) {
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

	authSvc := auth.New()
	result := &ImportResult{}

	for i, row := range rows[1:] { // Skip header
		rowNum := i + 2
		if len(row) < 2 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "用户名和密码为必填项",
			})
			continue
		}

		username := row[0]
		password := row[1]
		if username == "" || password == "" {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "用户名和密码不能为空",
			})
			continue
		}

		// Check username uniqueness
		cols := dao.SysUser.Columns()
		count, err := dao.SysUser.Ctx(ctx).
			Where(do.SysUser{Username: username}).
			WhereNull(cols.DeletedAt).
			Count()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("数据库查询错误: %v", err),
			})
			continue
		}
		if count > 0 {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("用户名 '%s' 已存在", username),
			})
			continue
		}

		hash, err := authSvc.HashPassword(password)
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: "密码加密失败",
			})
			continue
		}

		data := do.SysUser{
			Username:  username,
			Password:  hash,
			Status:    1,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}
		if len(row) > 2 {
			data.Nickname = row[2]
		}
		if len(row) > 3 {
			data.Phone = row[3]
		}
		if len(row) > 4 {
			data.Email = row[4]
		}
		if len(row) > 5 {
			switch row[5] {
			case "男", "1":
				data.Sex = 1
			case "女", "2":
				data.Sex = 2
			default:
				data.Sex = 0
			}
		}
		if len(row) > 6 {
			switch row[6] {
			case "停用", "0":
				data.Status = 0
			}
		}
		if len(row) > 7 {
			data.Remark = row[7]
		}

		_, err = dao.SysUser.Ctx(ctx).Data(data).Insert()
		if err != nil {
			result.Fail++
			result.FailList = append(result.FailList, ImportFailItem{
				Row:    rowNum,
				Reason: fmt.Sprintf("插入失败: %v", err),
			})
			continue
		}

		result.Success++
	}

	return result, nil
}

// GenerateImportTemplate creates an Excel template for user import.
func (s *Service) GenerateImportTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	headers := []string{"用户名", "密码", "昵称", "手机号码", "邮箱", "性别", "状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Example row
	f.SetCellValue(sheet, cellName(1, 2), "zhangsan")
	f.SetCellValue(sheet, cellName(2, 2), "123456")
	f.SetCellValue(sheet, cellName(3, 2), "张三")
	f.SetCellValue(sheet, cellName(4, 2), "13800138000")
	f.SetCellValue(sheet, cellName(5, 2), "zhangsan@example.com")
	f.SetCellValue(sheet, cellName(6, 2), "男")
	f.SetCellValue(sheet, cellName(7, 2), "正常")
	f.SetCellValue(sheet, cellName(8, 2), "示例用户")

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
