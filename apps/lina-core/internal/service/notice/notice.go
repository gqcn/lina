package notice

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/pkg/logger"
)

// Dict types used in notice
const (
	DictTypeNoticeType   = "sys_notice_type"   // Notice type dictionary
	DictTypeNoticeStatus = "sys_notice_status" // Notice status dictionary
)

// Notice type values (matching sys_notice_type dictionary)
const (
	NoticeTypeNotice       = 1 // 通知
	NoticeTypeAnnouncement = 2 // 公告
)

// Notice status values (matching sys_notice_status dictionary)
const (
	NoticeStatusDraft     = 0 // 草稿
	NoticeStatusPublished = 1 // 已发布
)

// Service provides notice management operations.
type Service struct {
	bizCtxSvc *bizctx.Service // Business context service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		bizCtxSvc: bizctx.New(),
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum   int    // Page number, starting from 1
	PageSize  int    // Page size
	Title     string // Title, supports fuzzy search
	Type      int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	CreatedBy string // Creator username, supports fuzzy search
}

// ListItem defines a single list item.
type ListItem struct {
	*entity.SysNotice        // Notice entity
	CreatedByName     string `json:"createdByName"` // Creator username
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListItem // List items
	Total int         // Total count
}

// List queries notice list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysNotice.Columns()
		m    = dao.SysNotice.Ctx(ctx)
	)

	// Apply filters
	if in.Title != "" {
		m = m.WhereLike(cols.Title, "%"+in.Title+"%")
	}
	if in.Type > 0 {
		m = m.Where(cols.Type, in.Type)
	}
	if in.CreatedBy != "" {
		// Filter by creator username via subquery
		userCols := dao.SysUser.Columns()
		subQuery := dao.SysUser.Ctx(ctx).
			Fields(userCols.Id).
			WhereLike(userCols.Username, "%"+in.CreatedBy+"%")
		m = m.Where(cols.CreatedBy+" IN (?)", subQuery)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysNotice
	err = m.Page(in.PageNum, in.PageSize).
		Order(cols.Id + " DESC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// Collect unique creator IDs
	userIds := make([]int64, 0, len(list))
	seen := make(map[int64]bool)
	for _, n := range list {
		if n.CreatedBy > 0 && !seen[n.CreatedBy] {
			userIds = append(userIds, n.CreatedBy)
			seen[n.CreatedBy] = true
		}
	}

	// Resolve creator usernames
	userNameMap := make(map[int64]string)
	if len(userIds) > 0 {
		var users []*entity.SysUser
		userCols := dao.SysUser.Columns()
		err = dao.SysUser.Ctx(ctx).
			Fields(userCols.Id, userCols.Username).
			WhereIn(userCols.Id, userIds).
			Scan(&users)
		if err == nil {
			for _, u := range users {
				userNameMap[int64(u.Id)] = u.Username
			}
		}
	}

	// Build result
	items := make([]*ListItem, 0, len(list))
	for _, n := range list {
		items = append(items, &ListItem{
			SysNotice:     n,
			CreatedByName: userNameMap[n.CreatedBy],
		})
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetById retrieves notice by ID.
func (s *Service) GetById(ctx context.Context, id int64) (*ListItem, error) {
	var notice *entity.SysNotice
	err := dao.SysNotice.Ctx(ctx).
		Where(do.SysNotice{Id: id}).
		Scan(&notice)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, gerror.New("通知公告不存在")
	}

	item := &ListItem{SysNotice: notice}

	// Resolve creator username
	if notice.CreatedBy > 0 {
		var user *entity.SysUser
		userCols := dao.SysUser.Columns()
		err = dao.SysUser.Ctx(ctx).
			Fields(userCols.Id, userCols.Username).
			Where(userCols.Id, notice.CreatedBy).
			Scan(&user)
		if err == nil && user != nil {
			item.CreatedByName = user.Username
		}
	}

	return item, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title   string // Title
	Type    int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	Content string // Content
	FileIds string // Attachment file IDs, comma-separated
	Status  int    // Status: 0=Draft 1=Published (see NoticeStatus* constants)
	Remark  string // Remark
}

// Create creates a new notice.
func (s *Service) Create(ctx context.Context, in CreateInput) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	var createdBy int64
	if bizCtx != nil {
		createdBy = int64(bizCtx.UserId)
	}

	// Insert notice (GoFrame auto-fills created_at and updated_at)
	id, err := dao.SysNotice.Ctx(ctx).Data(do.SysNotice{
		Title:     in.Title,
		Type:      in.Type,
		Content:   in.Content,
		FileIds:   in.FileIds,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	// If published, fan-out messages to all active users
	if in.Status == NoticeStatusPublished {
		if err := s.fanOutMessages(ctx, id, in.Title, in.Type, createdBy); err != nil {
			logger.Errorf(ctx, "fanOutMessages failed for notice %d: %v", id, err)
		}
	}

	return id, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id      int64   // Notice ID
	Title   *string // Title
	Type    *int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	Content *string // Content
	FileIds *string // Attachment file IDs, comma-separated
	Status  *int    // Status: 0=Draft 1=Published (see NoticeStatus* constants)
	Remark  *string // Remark
}

// Update updates notice information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check notice exists and get old status
	var oldNotice *entity.SysNotice
	err := dao.SysNotice.Ctx(ctx).
		Where(do.SysNotice{Id: in.Id}).
		Scan(&oldNotice)
	if err != nil {
		return err
	}
	if oldNotice == nil {
		return gerror.New("通知公告不存在")
	}

	bizCtx := s.bizCtxSvc.Get(ctx)
	var updatedBy int64
	if bizCtx != nil {
		updatedBy = int64(bizCtx.UserId)
	}

	data := do.SysNotice{
		UpdatedBy: updatedBy,
	}
	if in.Title != nil {
		data.Title = *in.Title
	}
	if in.Type != nil {
		data.Type = *in.Type
	}
	if in.Content != nil {
		data.Content = *in.Content
	}
	if in.FileIds != nil {
		data.FileIds = *in.FileIds
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err = dao.SysNotice.Ctx(ctx).Where(do.SysNotice{Id: in.Id}).Data(data).Update()
	if err != nil {
		return err
	}

	// If status changed from draft(0) to published(1), fan-out messages
	if in.Status != nil && *in.Status == NoticeStatusPublished && oldNotice.Status == NoticeStatusDraft {
		title := oldNotice.Title
		if in.Title != nil {
			title = *in.Title
		}
		noticeType := oldNotice.Type
		if in.Type != nil {
			noticeType = *in.Type
		}
		if err := s.fanOutMessages(ctx, in.Id, title, noticeType, int64(oldNotice.CreatedBy)); err != nil {
			logger.Errorf(ctx, "fanOutMessages failed for notice %d: %v", in.Id, err)
		}
	}

	return nil
}

// Delete soft-deletes notices by IDs and cascades to user messages.
func (s *Service) Delete(ctx context.Context, ids string) error {
	idList := strings.Split(ids, ",")
	if len(idList) == 0 {
		return gerror.New("请选择要删除的记录")
	}

	// Soft delete using GoFrame's auto soft-delete feature
	_, err := dao.SysNotice.Ctx(ctx).
		WhereIn(dao.SysNotice.Columns().Id, idList).
		Delete()
	if err != nil {
		return err
	}

	// Cascade delete corresponding user messages
	msgCols := dao.SysUserMessage.Columns()
	_, err = dao.SysUserMessage.Ctx(ctx).
		Where(msgCols.SourceType, "notice").
		WhereIn(msgCols.SourceId, idList).
		Delete()
	if err != nil {
		logger.Errorf(ctx, "cascade delete user messages failed for notice ids %s: %v", ids, err)
	}
	return nil
}

// fanOutMessages creates user_message records for all active users.
func (s *Service) fanOutMessages(ctx context.Context, noticeId int64, title string, noticeType int, createdBy int64) error {
	var users []*entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Status: 1}).
		WhereNot(dao.SysUser.Columns().Id, createdBy).
		Scan(&users)
	if err != nil {
		return err
	}

	for _, user := range users {
		_, err = dao.SysUserMessage.Ctx(ctx).Data(do.SysUserMessage{
			UserId:     user.Id,
			Title:      title,
			Type:       noticeType,
			SourceType: "notice",
			SourceId:   noticeId,
			IsRead:     0,
		}).Insert()
		if err != nil {
			logger.Errorf(ctx, "fanOutMessages insert failed for user %d notice %d: %v", user.Id, noticeId, err)
		}
	}
	logger.Infof(ctx, "fanOutMessages: notice %d fanned out to %d users", noticeId, len(users))
	return nil
}
