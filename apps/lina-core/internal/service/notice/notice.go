package notice

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
)

// Service provides notice management operations.
type Service struct {
	bizCtxSvc *bizctx.Service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		bizCtxSvc: bizctx.New(),
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum   int
	PageSize  int
	Title     string
	Type      int
	CreatedBy string
}

// ListItem defines a single list item with creator name.
type ListItem struct {
	*entity.SysNotice
	CreatedByName string `json:"createdByName"`
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListItem
	Total int
}

// List queries notice list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysNotice.Columns()
		m    = dao.SysNotice.Ctx(ctx).WhereNull(cols.DeletedAt)
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

	// Build result with creator names
	items := make([]*ListItem, 0, len(list))
	for _, n := range list {
		item := &ListItem{SysNotice: n}
		if n.CreatedBy > 0 {
			var user *entity.SysUser
			_ = dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: int(n.CreatedBy)}).Scan(&user)
			if user != nil {
				item.CreatedByName = user.Nickname
				if item.CreatedByName == "" {
					item.CreatedByName = user.Username
				}
			}
		}
		items = append(items, item)
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetById retrieves notice by ID.
func (s *Service) GetById(ctx context.Context, id int64) (*ListItem, error) {
	var notice *entity.SysNotice
	cols := dao.SysNotice.Columns()
	err := dao.SysNotice.Ctx(ctx).
		Where(do.SysNotice{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&notice)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, gerror.New("通知公告不存在")
	}

	item := &ListItem{SysNotice: notice}
	if notice.CreatedBy > 0 {
		var user *entity.SysUser
		_ = dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: int(notice.CreatedBy)}).Scan(&user)
		if user != nil {
			item.CreatedByName = user.Nickname
			if item.CreatedByName == "" {
				item.CreatedByName = user.Username
			}
		}
	}
	return item, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title   string
	Type    int
	Content string
	Status  int
	Remark  string
}

// Create creates a new notice.
func (s *Service) Create(ctx context.Context, in CreateInput) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	var createdBy int64
	if bizCtx != nil {
		createdBy = int64(bizCtx.UserId)
	}

	id, err := dao.SysNotice.Ctx(ctx).Data(do.SysNotice{
		Title:     in.Title,
		Type:      in.Type,
		Content:   in.Content,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	// If published, fan-out messages to all active users
	if in.Status == 1 {
		_ = s.fanOutMessages(ctx, id, in.Title, in.Type, createdBy)
	}

	return id, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id      int64
	Title   *string
	Type    *int
	Content *string
	Status  *int
	Remark  *string
}

// Update updates notice information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check notice exists and get old status
	cols := dao.SysNotice.Columns()
	var oldNotice *entity.SysNotice
	err := dao.SysNotice.Ctx(ctx).
		Where(do.SysNotice{Id: in.Id}).
		WhereNull(cols.DeletedAt).
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
		UpdatedAt: gtime.Now(),
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
	if in.Status != nil && *in.Status == 1 && oldNotice.Status == 0 {
		title := oldNotice.Title
		if in.Title != nil {
			title = *in.Title
		}
		noticeType := oldNotice.Type
		if in.Type != nil {
			noticeType = *in.Type
		}
		_ = s.fanOutMessages(ctx, in.Id, title, noticeType, int64(oldNotice.CreatedBy))
	}

	return nil
}

// Delete soft-deletes notices by IDs.
func (s *Service) Delete(ctx context.Context, ids string) error {
	idList := strings.Split(ids, ",")
	if len(idList) == 0 {
		return gerror.New("请选择要删除的记录")
	}

	_, err := dao.SysNotice.Ctx(ctx).
		WhereIn(dao.SysNotice.Columns().Id, idList).
		Data(do.SysNotice{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// fanOutMessages creates user_message records for all active users.
func (s *Service) fanOutMessages(ctx context.Context, noticeId int64, title string, noticeType int, createdBy int64) error {
	userCols := dao.SysUser.Columns()
	var users []*entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Status: 1}).
		WhereNull(userCols.DeletedAt).
		WhereNot(userCols.Id, createdBy).
		Scan(&users)
	if err != nil {
		return err
	}

	for _, user := range users {
		_, _ = dao.SysUserMessage.Ctx(ctx).Data(do.SysUserMessage{
			UserId:     user.Id,
			Title:      title,
			Type:       noticeType,
			SourceType: "notice",
			SourceId:   noticeId,
			IsRead:     0,
			CreatedAt:  gtime.Now(),
		}).Insert()
	}
	return nil
}
