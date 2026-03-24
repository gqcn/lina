package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
)

// Service provides user message operations.
type Service struct {
	bizCtxSvc *bizctx.Service // 业务上下文服务
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		bizCtxSvc: bizctx.New(),
	}
}

// getCurrentUserId extracts current user ID from context.
func (s *Service) getCurrentUserId(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId == 0 {
		return 0, gerror.New("未登录")
	}
	return int64(bizCtx.UserId), nil
}

// UnreadCount returns unread message count for current user.
func (s *Service) UnreadCount(ctx context.Context) (int, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return 0, err
	}

	cols := dao.SysUserMessage.Columns()
	count, err := dao.SysUserMessage.Ctx(ctx).
		Where(cols.UserId, userId).
		Where(cols.IsRead, 0).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum  int // 页码，从1开始
	PageSize int // 每页数量
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysUserMessage // 消息列表
	Total int                      // 总数
}

// List queries message list for current user with pagination.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return nil, err
	}

	cols := dao.SysUserMessage.Columns()
	m := dao.SysUserMessage.Ctx(ctx).Where(do.SysUserMessage{UserId: userId})

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var list []*entity.SysUserMessage
	err = m.Page(in.PageNum, in.PageSize).
		Order(cols.Id + " DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{List: list, Total: total}, nil
}

// MarkRead marks a single message as read.
func (s *Service) MarkRead(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}

	_, err = dao.SysUserMessage.Ctx(ctx).
		Where(do.SysUserMessage{Id: id, UserId: userId}).
		Data(do.SysUserMessage{IsRead: 1, ReadAt: gtime.Now()}).
		Update()
	return err
}

// MarkReadAll marks all messages as read for current user.
func (s *Service) MarkReadAll(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}

	cols := dao.SysUserMessage.Columns()
	_, err = dao.SysUserMessage.Ctx(ctx).
		Where(cols.UserId, userId).
		Where(cols.IsRead, 0).
		Data(do.SysUserMessage{IsRead: 1, ReadAt: gtime.Now()}).
		Update()
	return err
}

// Delete physically deletes a single message.
func (s *Service) Delete(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}

	_, err = dao.SysUserMessage.Ctx(ctx).
		Where(do.SysUserMessage{Id: id, UserId: userId}).
		Delete()
	return err
}

// Clear physically deletes all messages for current user.
func (s *Service) Clear(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}

	_, err = dao.SysUserMessage.Ctx(ctx).
		Where(do.SysUserMessage{UserId: userId}).
		Delete()
	return err
}
