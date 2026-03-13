package user

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"backend/internal/consts"
	"backend/internal/dao"
	"backend/internal/model/do"
	"backend/internal/model/entity"
	"backend/internal/service/auth"
	"backend/internal/service/bizctx"
)

// Service provides user management operations.
type Service struct {
	authSvc   *auth.Service
	bizCtxSvc *bizctx.Service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		authSvc:   auth.New(),
		bizCtxSvc: bizctx.New(),
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int
	PageSize       int
	Username       string
	Nickname       string
	Status         *int
	Phone          string
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysUser
	Total int
}

// List queries user list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysUser.Columns()
		m    = dao.SysUser.Ctx(ctx).WhereNull(cols.DeletedAt)
	)

	// Apply filters
	if in.Username != "" {
		m = m.WhereLike(cols.Username, "%"+in.Username+"%")
	}
	if in.Nickname != "" {
		m = m.WhereLike(cols.Nickname, "%"+in.Nickname+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.Phone != "" {
		m = m.WhereLike(cols.Phone, "%"+in.Phone+"%")
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Determine sort order
	allowedSortFields := map[string]string{
		"id":         cols.Id,
		"username":   cols.Username,
		"nickname":   cols.Nickname,
		"phone":      cols.Phone,
		"email":      cols.Email,
		"status":     cols.Status,
		"created_at": cols.CreatedAt,
		"createdAt":  cols.CreatedAt,
	}
	sortField := cols.Id
	if f, ok := allowedSortFields[in.OrderBy]; ok {
		sortField = f
	}
	sortDirection := "DESC"
	if in.OrderDirection == "asc" {
		sortDirection = "ASC"
	}

	// Query with pagination, exclude password field
	var list []*entity.SysUser
	err = m.FieldsEx(cols.Password).
		Page(in.PageNum, in.PageSize).
		Order(sortField + " " + sortDirection).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Status   int
	Remark   string
}

// Create creates a new user.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check username uniqueness
	cols := dao.SysUser.Columns()
	count, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
		WhereNull(cols.DeletedAt).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("用户名已存在")
	}

	// Hash password
	hash, err := s.authSvc.HashPassword(in.Password)
	if err != nil {
		return 0, err
	}

	// Insert user
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username:  in.Username,
		Password:  hash,
		Nickname:  in.Nickname,
		Email:     in.Email,
		Phone:     in.Phone,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetById retrieves user by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysUser, error) {
	var user *entity.SysUser
	cols := dao.SysUser.Columns()
	err := dao.SysUser.Ctx(ctx).
		FieldsEx(cols.Password).
		Where(do.SysUser{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, gerror.New("用户不存在")
	}
	return user, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id       int
	Username *string
	Password *string
	Nickname *string
	Email    *string
	Phone    *string
	Status   *int
	Remark   *string
}

// Update updates user information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Cannot edit self via admin panel
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == in.Id {
		return gerror.New("不能编辑当前登录用户")
	}

	// Check user exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysUser{
		UpdatedAt: gtime.Now(),
	}
	if in.Username != nil {
		data.Username = *in.Username
	}
	if in.Password != nil && *in.Password != "" {
		hash, err := s.authSvc.HashPassword(*in.Password)
		if err != nil {
			return err
		}
		data.Password = hash
	}
	if in.Nickname != nil {
		data.Nickname = *in.Nickname
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a user.
func (s *Service) Delete(ctx context.Context, id int) error {
	// Cannot delete default admin
	if id == consts.DefaultAdminId {
		return gerror.New("不能删除默认管理员")
	}

	// Cannot delete self
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id {
		return gerror.New("不能删除当前登录用户")
	}

	// Soft delete
	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Data(do.SysUser{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// UpdateStatus updates user status.
func (s *Service) UpdateStatus(ctx context.Context, id int, status int) error {
	// Cannot disable self
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx != nil && bizCtx.UserId == id && status == consts.UserStatusDisabled {
		return gerror.New("不能停用当前登录用户")
	}

	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Data(do.SysUser{
			Status:    status,
			UpdatedAt: gtime.Now(),
		}).
		Update()
	return err
}

// GetProfile retrieves current user profile.
func (s *Service) GetProfile(ctx context.Context) (*entity.SysUser, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return nil, gerror.New("未登录")
	}
	return s.GetById(ctx, bizCtx.UserId)
}

// UpdateProfileInput defines input for UpdateProfile function.
type UpdateProfileInput struct {
	Nickname *string
	Email    *string
	Phone    *string
	Password *string
}

// UpdateProfile updates current user profile.
func (s *Service) UpdateProfile(ctx context.Context, in UpdateProfileInput) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return gerror.New("未登录")
	}

	data := do.SysUser{
		UpdatedAt: gtime.Now(),
	}
	if in.Nickname != nil {
		data.Nickname = *in.Nickname
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Password != nil && *in.Password != "" {
		hash, err := s.authSvc.HashPassword(*in.Password)
		if err != nil {
			return err
		}
		data.Password = hash
	}

	_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: bizCtx.UserId}).Data(data).Update()
	return err
}

// ResetPassword resets a user's password.
func (s *Service) ResetPassword(ctx context.Context, id int, password string) error {
	// Check user exists
	if _, err := s.GetById(ctx, id); err != nil {
		return err
	}

	// Hash password
	hash, err := s.authSvc.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Data(do.SysUser{
			Password:  hash,
			UpdatedAt: gtime.Now(),
		}).
		Update()
	return err
}

// UpdateAvatar updates current user's avatar URL.
func (s *Service) UpdateAvatar(ctx context.Context, avatarUrl string) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return gerror.New("未登录")
	}
	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: bizCtx.UserId}).
		Data(do.SysUser{
			Avatar:    avatarUrl,
			UpdatedAt: gtime.Now(),
		}).
		Update()
	return err
}
