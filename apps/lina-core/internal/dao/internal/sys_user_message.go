// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserMessageDao is the data access object for the table sys_user_message.
type SysUserMessageDao struct {
	table    string                // table is the underlying table name of the DAO.
	group    string                // group is the database configuration group name of the current DAO.
	columns  SysUserMessageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler    // handlers for customized model modification.
}

// SysUserMessageColumns defines and stores column names for the table sys_user_message.
type SysUserMessageColumns struct {
	Id         string // 消息ID
	UserId     string // 接收用户ID
	Title      string // 消息标题
	Type       string // 消息类型（1通知 2公告）
	SourceType string // 来源类型
	SourceId   string // 来源ID
	IsRead     string // 是否已读（0未读 1已读）
	ReadAt     string // 阅读时间
	CreatedAt  string // 创建时间
}

// sysUserMessageColumns holds the columns for the table sys_user_message.
var sysUserMessageColumns = SysUserMessageColumns{
	Id:         "id",
	UserId:     "user_id",
	Title:      "title",
	Type:       "type",
	SourceType: "source_type",
	SourceId:   "source_id",
	IsRead:     "is_read",
	ReadAt:     "read_at",
	CreatedAt:  "created_at",
}

// NewSysUserMessageDao creates and returns a new DAO object for table data access.
func NewSysUserMessageDao(handlers ...gdb.ModelHandler) *SysUserMessageDao {
	return &SysUserMessageDao{
		group:    "default",
		table:    "sys_user_message",
		columns:  sysUserMessageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SysUserMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SysUserMessageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SysUserMessageDao) Columns() SysUserMessageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SysUserMessageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SysUserMessageDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SysUserMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
