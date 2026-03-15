// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysOperLogDao is the data access object for the table sys_oper_log.
type SysOperLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SysOperLogColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SysOperLogColumns defines and stores column names for the table sys_oper_log.
type SysOperLogColumns struct {
	Id            string //
	Title         string //
	OperSummary   string //
	OperType      string //
	Method        string //
	RequestMethod string //
	OperName      string //
	OperUrl       string //
	OperIp        string //
	OperParam     string //
	JsonResult    string //
	Status        string //
	ErrorMsg      string //
	CostTime      string //
	OperTime      string //
}

// sysOperLogColumns holds the columns for the table sys_oper_log.
var sysOperLogColumns = SysOperLogColumns{
	Id:            "id",
	Title:         "title",
	OperSummary:   "oper_summary",
	OperType:      "oper_type",
	Method:        "method",
	RequestMethod: "request_method",
	OperName:      "oper_name",
	OperUrl:       "oper_url",
	OperIp:        "oper_ip",
	OperParam:     "oper_param",
	JsonResult:    "json_result",
	Status:        "status",
	ErrorMsg:      "error_msg",
	CostTime:      "cost_time",
	OperTime:      "oper_time",
}

// NewSysOperLogDao creates and returns a new DAO object for table data access.
func NewSysOperLogDao(handlers ...gdb.ModelHandler) *SysOperLogDao {
	return &SysOperLogDao{
		group:    "default",
		table:    "sys_oper_log",
		columns:  sysOperLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysOperLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysOperLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysOperLogDao) Columns() SysOperLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysOperLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysOperLogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SysOperLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
