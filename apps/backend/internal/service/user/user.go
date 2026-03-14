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
	Sex            *int
	DeptId         *int
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
}

// ListOutputItem defines a single item in list output with dept info.
type ListOutputItem struct {
	SysUser  *entity.SysUser
	DeptId   int
	DeptName string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListOutputItem
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
	if in.Sex != nil {
		m = m.Where(cols.Sex, *in.Sex)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime)
	}

	// Filter by dept via association table
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			// Unassigned: users NOT in sys_user_dept
			assignedUserIds, err := s.GetAllAssignedUserIds(ctx)
			if err != nil {
				return nil, err
			}
			if len(assignedUserIds) > 0 {
				m = m.WhereNotIn(cols.Id, assignedUserIds)
			}
		} else {
			userIds, err := s.GetUserIdsByDeptId(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			if len(userIds) == 0 {
				return &ListOutput{List: []*ListOutputItem{}, Total: 0}, nil
			}
			m = m.WhereIn(cols.Id, userIds)
		}
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

	// Build output with dept info
	items := make([]*ListOutputItem, 0, len(list))
	for _, u := range list {
		item := &ListOutputItem{SysUser: u}
		// Get dept info from association table
		deptId, deptName, _ := s.GetUserDeptInfo(ctx, u.Id)
		item.DeptId = deptId
		item.DeptName = deptName
		items = append(items, item)
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetUserIdsByDeptId returns user IDs associated with a dept.
func (s *Service) GetUserIdsByDeptId(ctx context.Context, deptId int) ([]int, error) {
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Where(dao.SysUserDept.Columns().DeptId, deptId).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		ids = append(ids, ud.UserId)
	}
	return ids, nil
}

// GetAllAssignedUserIds returns all user IDs that have a dept association.
func (s *Service) GetAllAssignedUserIds(ctx context.Context) ([]int, error) {
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Fields(dao.SysUserDept.Columns().UserId).
		Distinct().
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		ids = append(ids, ud.UserId)
	}
	return ids, nil
}

// GetUserDeptInfo returns the dept ID and name for a user.
func (s *Service) GetUserDeptInfo(ctx context.Context, userId int) (int, string, error) {
	var userDept *entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Where(dao.SysUserDept.Columns().UserId, userId).
		Scan(&userDept)
	if err != nil || userDept == nil {
		return 0, "", err
	}
	var dept *entity.SysDept
	deptCols := dao.SysDept.Columns()
	err = dao.SysDept.Ctx(ctx).
		Where(dao.SysDept.Columns().Id, userDept.DeptId).
		WhereNull(deptCols.DeletedAt).
		Scan(&dept)
	if err != nil || dept == nil {
		return 0, "", err
	}
	return dept.Id, dept.Name, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Sex      int
	Status   int
	Remark   string
	DeptId   *int
	PostIds  []int
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
		Sex:       in.Sex,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	userId := int(id)

	// Save dept association
	if in.DeptId != nil && *in.DeptId > 0 {
		_, err = dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{
			UserId: userId,
			DeptId: *in.DeptId,
		}).Insert()
		if err != nil {
			return 0, err
		}
	}

	// Save post associations
	for _, postId := range in.PostIds {
		_, err = dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{
			UserId: userId,
			PostId: postId,
		}).Insert()
		if err != nil {
			return 0, err
		}
	}

	return userId, nil
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
	Sex      *int
	Status   *int
	Remark   *string
	DeptId   *int
	PostIds  []int
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
	if in.Sex != nil {
		data.Sex = *in.Sex
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: in.Id}).Data(data).Update()
	if err != nil {
		return err
	}

	// Update dept association (delete and re-insert)
	if in.DeptId != nil {
		_, _ = dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, in.Id).Delete()
		if *in.DeptId > 0 {
			_, err = dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{
				UserId: in.Id,
				DeptId: *in.DeptId,
			}).Insert()
			if err != nil {
				return err
			}
		}
	}

	// Update post associations (delete and re-insert)
	if in.PostIds != nil {
		_, _ = dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, in.Id).Delete()
		for _, postId := range in.PostIds {
			_, err = dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{
				UserId: in.Id,
				PostId: postId,
			}).Insert()
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	if err != nil {
		return err
	}

	// Clean up dept and post associations
	_, _ = dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, id).Delete()
	_, _ = dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, id).Delete()

	return nil
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
	Sex      *int
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
	if in.Sex != nil {
		data.Sex = *in.Sex
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

// GetUserPostIds returns the post IDs associated with a user.
func (s *Service) GetUserPostIds(ctx context.Context, userId int) ([]int, error) {
	var userPosts []*entity.SysUserPost
	err := dao.SysUserPost.Ctx(ctx).
		Where(dao.SysUserPost.Columns().UserId, userId).
		Scan(&userPosts)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userPosts))
	for _, up := range userPosts {
		ids = append(ids, up.PostId)
	}
	return ids, nil
}
