package user

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/consts"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/dept"
	"lina-core/internal/service/role"
)

// Service provides user management operations.
type Service struct {
	authSvc   *auth.Service   // Authentication service
	bizCtxSvc *bizctx.Service // Business context service
	deptSvc   *dept.Service   // Department service
	roleSvc   *role.Service   // Role service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		authSvc:   auth.New(),
		bizCtxSvc: bizctx.New(),
		deptSvc:   dept.New(),
		roleSvc:   role.New(),
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum        int    // Page number, starting from 1
	PageSize       int    // Items per page
	Username       string // Username, supports fuzzy search
	Nickname       string // Nickname, supports fuzzy search
	Status         *int   // Status: 1=Normal 0=Disabled
	Phone          string // Phone number, supports fuzzy search
	Sex            *int   // Gender: 0=Unknown 1=Male 2=Female
	DeptId         *int   // Department ID, 0 means unassigned
	BeginTime      string // Creation time start
	EndTime        string // Creation time end
	OrderBy        string // Sort field
	OrderDirection string // Sort direction: asc/desc
}

// ListOutputItem defines a single item in list output with dept info.
type ListOutputItem struct {
	SysUser   *entity.SysUser // User entity
	DeptId    int             // Department ID
	DeptName  string          // Department name
	RoleIds   []int           // Role ID list
	RoleNames []string        // Role name list
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListOutputItem // User list
	Total int               // Total count
}

// List queries user list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysUser.Columns()
		m    = dao.SysUser.Ctx(ctx)
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

	// Batch query dept info to avoid N+1 problem
	items := make([]*ListOutputItem, 0, len(list))
	if len(list) == 0 {
		return &ListOutput{List: items, Total: total}, nil
	}

	// Collect all user IDs
	userIds := make([]int, 0, len(list))
	for _, u := range list {
		userIds = append(userIds, u.Id)
	}

	// Batch query user-dept associations
	udCols := dao.SysUserDept.Columns()
	var userDepts []*entity.SysUserDept
	err = dao.SysUserDept.Ctx(ctx).
		WhereIn(udCols.UserId, userIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	// Build userId -> deptId map
	userDeptMap := make(map[int]int)
	deptIds := make([]int, 0)
	for _, ud := range userDepts {
		userDeptMap[ud.UserId] = ud.DeptId
		deptIds = append(deptIds, ud.DeptId)
	}

	// Batch query dept info
	deptCols := dao.SysDept.Columns()
	var depts []*entity.SysDept
	if len(deptIds) > 0 {
		err = dao.SysDept.Ctx(ctx).
			WhereIn(deptCols.Id, deptIds).
			WhereNull(deptCols.DeletedAt).
			Scan(&depts)
		if err != nil {
			return nil, err
		}
	}

	// Build deptId -> deptName map
	deptNameMap := make(map[int]string)
	for _, d := range depts {
		deptNameMap[d.Id] = d.Name
	}

	// Build user-role associations
	urCols := dao.SysUserRole.Columns()
	var userRoles []*entity.SysUserRole
	err = dao.SysUserRole.Ctx(ctx).
		WhereIn(urCols.UserId, userIds).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}

	// Build userId -> roleIds map
	userRoleMap := make(map[int][]int)
	roleIdsSet := make(map[int]bool)
	for _, ur := range userRoles {
		userRoleMap[ur.UserId] = append(userRoleMap[ur.UserId], ur.RoleId)
		roleIdsSet[ur.RoleId] = true
	}

	// Get all unique role IDs
	allRoleIds := make([]int, 0, len(roleIdsSet))
	for roleId := range roleIdsSet {
		allRoleIds = append(allRoleIds, roleId)
	}

	// Batch query role info
	roleCols := dao.SysRole.Columns()
	var roles []*entity.SysRole
	if len(allRoleIds) > 0 {
		err = dao.SysRole.Ctx(ctx).
			WhereIn(roleCols.Id, allRoleIds).
			Scan(&roles)
		if err != nil {
			return nil, err
		}
	}

	// Build roleId -> roleName map
	roleNameMap := make(map[int]string)
	for _, r := range roles {
		roleNameMap[r.Id] = r.Name
	}

	// Build output with dept and role info
	for _, u := range list {
		item := &ListOutputItem{SysUser: u}
		if deptId, ok := userDeptMap[u.Id]; ok {
			item.DeptId = deptId
			item.DeptName = deptNameMap[deptId]
		}
		// Get role info
		if roleIds, ok := userRoleMap[u.Id]; ok {
			item.RoleIds = roleIds
			for _, roleId := range roleIds {
				if name, exists := roleNameMap[roleId]; exists {
					item.RoleNames = append(item.RoleNames, name)
				}
			}
		} else {
			item.RoleIds = []int{}
			item.RoleNames = []string{}
		}
		items = append(items, item)
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetUserIdsByDeptId returns user IDs associated with a dept and all its descendants.
func (s *Service) GetUserIdsByDeptId(ctx context.Context, deptId int) ([]int, error) {
	// Use shared method from dept service to get dept and descendant IDs
	deptIds, err := s.deptSvc.GetDeptAndDescendantIds(ctx, deptId)
	if err != nil {
		return nil, err
	}

	// Query users belonging to any of these depts
	var userDepts []*entity.SysUserDept
	err = dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().DeptId, deptIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}
	// Deduplicate user IDs (a user could belong to multiple depts in the subtree)
	seen := make(map[int]struct{})
	ids := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		if _, ok := seen[ud.UserId]; !ok {
			seen[ud.UserId] = struct{}{}
			ids = append(ids, ud.UserId)
		}
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
	Username string // Username
	Password string // Password
	Nickname string // Nickname
	Email    string // Email
	Phone    string // Phone number
	Sex      int    // Gender: 0=Unknown 1=Male 2=Female
	Status   int    // Status: 1=Normal 0=Disabled
	Remark   string // Remark
	DeptId   *int   // Department ID
	PostIds  []int  // Post ID list
	RoleIds  []int  // Role ID list
}

// Create creates a new user with transaction support.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check username uniqueness
	count, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
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

	// Default nickname to username if empty
	nickname := in.Nickname
	if nickname == "" {
		nickname = in.Username
	}

	var userId int

	// Use transaction to ensure atomicity
	err = dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Insert user (GoFrame auto-fills created_at and updated_at)
		id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
			Username: in.Username,
			Password: hash,
			Nickname: nickname,
			Email:    in.Email,
			Phone:    in.Phone,
			Sex:      in.Sex,
			Status:   in.Status,
			Remark:   in.Remark,
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		userId = int(id)

		// Save dept association
		if in.DeptId != nil && *in.DeptId > 0 {
			_, err = dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{
				UserId: userId,
				DeptId: *in.DeptId,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Save post associations
		for _, postId := range in.PostIds {
			_, err = dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{
				UserId: userId,
				PostId: postId,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Save role associations
		for _, roleId := range in.RoleIds {
			_, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
				UserId: userId,
				RoleId: roleId,
			}).Insert()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
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
	Id       int      // User ID
	Username *string  // Username
	Password *string  // Password
	Nickname *string  // Nickname
	Email    *string  // Email
	Phone    *string  // Phone number
	Sex      *int     // Gender: 0=Unknown 1=Male 2=Female
	Status   *int     // Status: 1=Normal 0=Disabled
	Remark   *string  // Remark
	DeptId   *int     // Department ID
	PostIds  []int    // Post ID list
	RoleIds  []int    // Role ID list
}

// Update updates user information with transaction support.
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

	data := do.SysUser{}
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

	// Use transaction to ensure atomicity
	return dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Update user
		_, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		// Update dept association (delete and re-insert)
		if in.DeptId != nil {
			_, err = dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, in.Id).Delete()
			if err != nil {
				g.Log().Warningf(ctx, "failed to delete user dept association: %v", err)
			}
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
			_, err = dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, in.Id).Delete()
			if err != nil {
				g.Log().Warningf(ctx, "failed to delete user post association: %v", err)
			}
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

		// Update role associations (delete and re-insert)
		if in.RoleIds != nil {
			_, err = dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().UserId, in.Id).Delete()
			if err != nil {
				g.Log().Warningf(ctx, "failed to delete user role association: %v", err)
			}
			for _, roleId := range in.RoleIds {
				_, err = dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
					UserId: in.Id,
					RoleId: roleId,
				}).Insert()
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
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

	// Soft delete using GoFrame's auto soft-delete feature
	_, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id}).
		Delete()
	if err != nil {
		return err
	}

	// Clean up dept, post and role associations (log errors but don't fail)
	if _, err := dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, id).Delete(); err != nil {
		g.Log().Warningf(ctx, "failed to delete user dept association for user %d: %v", id, err)
	}
	if _, err := dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, id).Delete(); err != nil {
		g.Log().Warningf(ctx, "failed to delete user post association for user %d: %v", id, err)
	}
	if _, err := dao.SysUserRole.Ctx(ctx).Where(dao.SysUserRole.Columns().UserId, id).Delete(); err != nil {
		g.Log().Warningf(ctx, "failed to delete user role association for user %d: %v", id, err)
	}

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
			Status: status,
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
	Nickname *string // Nickname
	Email    *string // Email
	Phone    *string // Phone number
	Sex      *int    // Gender: 0=Unknown 1=Male 2=Female
	Password *string // Password
}

// UpdateProfile updates current user profile.
func (s *Service) UpdateProfile(ctx context.Context, in UpdateProfileInput) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil {
		return gerror.New("未登录")
	}

	data := do.SysUser{}
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
			Password: hash,
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
			Avatar: avatarUrl,
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

// GetUserRoleIds returns the role IDs associated with a user.
func (s *Service) GetUserRoleIds(ctx context.Context, userId int) ([]int, error) {
	var userRoles []*entity.SysUserRole
	err := dao.SysUserRole.Ctx(ctx).
		Where(dao.SysUserRole.Columns().UserId, userId).
		Scan(&userRoles)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(userRoles))
	for _, ur := range userRoles {
		ids = append(ids, ur.RoleId)
	}
	return ids, nil
}
