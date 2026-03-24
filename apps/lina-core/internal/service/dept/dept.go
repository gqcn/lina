package dept

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Service provides dept management operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// TreeNode defines the tree structure for dept.
type TreeNode struct {
	Id        int         `json:"id"`        // 部门ID
	Label     string      `json:"label"`     // 部门名称（含用户数）
	UserCount int         `json:"userCount"` // 用户数量
	Children  []*TreeNode `json:"children"`  // 子部门列表
}

// DeptUser defines the user info in a dept.
type DeptUser struct {
	Id       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
}

// ListInput defines input for List function.
type ListInput struct {
	Name   string // 部门名称，支持模糊查询
	Status *int   // 状态：1=正常 0=停用
}

// ListOutput defines output for List function.
type ListOutput struct {
	List []*entity.SysDept // 部门列表
}

// List queries dept list with filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysDept.Columns()
		m    = dao.SysDept.Ctx(ctx).WhereNull(cols.DeletedAt)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	// Query all, ordered by order_num ASC
	var list []*entity.SysDept
	err := m.Order(cols.OrderNum + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List: list,
	}, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	ParentId int    // 父部门ID，0表示顶级部门
	Name     string // 部门名称
	Code     string // 部门编码
	OrderNum int    // 显示顺序
	Leader   int    // 负责人用户ID
	Phone    string // 联系电话
	Email    string // 邮箱
	Status   int    // 状态：1=正常 0=停用
	Remark   string // 备注
}

// Create creates a new dept.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check code uniqueness
	if in.Code != "" {
		if err := s.checkCodeUnique(ctx, in.Code, 0); err != nil {
			return 0, err
		}
	}

	// Calculate ancestors
	var ancestors string
	if in.ParentId == 0 {
		ancestors = "0"
	} else {
		parent, err := s.GetById(ctx, in.ParentId)
		if err != nil {
			return 0, err
		}
		ancestors = fmt.Sprintf("%s,%d", parent.Ancestors, in.ParentId)
	}

	// Insert dept
	id, err := dao.SysDept.Ctx(ctx).Data(do.SysDept{
		ParentId:  in.ParentId,
		Ancestors: ancestors,
		Name:      in.Name,
		Code:      in.Code,
		OrderNum:  in.OrderNum,
		Leader:    in.Leader,
		Phone:     in.Phone,
		Email:     in.Email,
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

// GetById retrieves dept by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysDept, error) {
	var dept *entity.SysDept
	cols := dao.SysDept.Columns()
	err := dao.SysDept.Ctx(ctx).
		Where(do.SysDept{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&dept)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, gerror.New("部门不存在")
	}
	return dept, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id       int     // 部门ID
	ParentId *int    // 父部门ID
	Name     *string // 部门名称
	Code     *string // 部门编码
	OrderNum *int    // 显示顺序
	Leader   *int    // 负责人用户ID
	Phone    *string // 联系电话
	Email    *string // 邮箱
	Status   *int    // 状态：1=正常 0=停用
	Remark   *string // 备注
}

// Update updates dept information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check dept exists
	dept, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}

	data := do.SysDept{
		UpdatedAt: gtime.Now(),
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Code != nil {
		if *in.Code != "" {
			if err := s.checkCodeUnique(ctx, *in.Code, in.Id); err != nil {
				return err
			}
		}
		data.Code = *in.Code
	}
	if in.OrderNum != nil {
		data.OrderNum = *in.OrderNum
	}
	if in.Leader != nil {
		data.Leader = *in.Leader
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	// Handle parent change: recalculate ancestors
	if in.ParentId != nil && *in.ParentId != dept.ParentId {
		newParentId := *in.ParentId
		var newAncestors string
		if newParentId == 0 {
			newAncestors = "0"
		} else {
			parent, err := s.GetById(ctx, newParentId)
			if err != nil {
				return err
			}
			newAncestors = fmt.Sprintf("%s,%d", parent.Ancestors, newParentId)
		}

		oldAncestors := dept.Ancestors
		data.ParentId = newParentId
		data.Ancestors = newAncestors

		// Update children's ancestors
		oldPrefix := fmt.Sprintf("%s,%d", oldAncestors, in.Id)
		newPrefix := fmt.Sprintf("%s,%d", newAncestors, in.Id)

		cols := dao.SysDept.Columns()
		var children []*entity.SysDept
		err = dao.SysDept.Ctx(ctx).
			WhereNull(cols.DeletedAt).
			Where(
				dao.SysDept.Ctx(ctx).Builder().
					WhereLike(cols.Ancestors, oldPrefix+",%").
					WhereOr(cols.ParentId, in.Id),
			).
			Scan(&children)
		if err != nil {
			return err
		}

		for _, child := range children {
			childNewAncestors := gstr.Replace(child.Ancestors, oldPrefix, newPrefix, 1)
			_, err = dao.SysDept.Ctx(ctx).
				Where(do.SysDept{Id: child.Id}).
				Data(do.SysDept{
					Ancestors: childNewAncestors,
					UpdatedAt: gtime.Now(),
				}).
				Update()
			if err != nil {
				return err
			}
		}
	}

	_, err = dao.SysDept.Ctx(ctx).Where(do.SysDept{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a dept.
func (s *Service) Delete(ctx context.Context, id int) error {
	cols := dao.SysDept.Columns()

	// Check no children
	childCount, err := dao.SysDept.Ctx(ctx).
		Where(cols.ParentId, id).
		WhereNull(cols.DeletedAt).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return gerror.New("存在子部门，不允许删除")
	}

	// Check no users in dept
	userCount, err := dao.SysUserDept.Ctx(ctx).
		Where(do.SysUserDept{DeptId: id}).
		Count()
	if err != nil {
		return err
	}
	if userCount > 0 {
		return gerror.New("部门存在用户，不允许删除")
	}

	// Soft delete
	_, err = dao.SysDept.Ctx(ctx).
		Where(do.SysDept{Id: id}).
		Data(do.SysDept{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// Tree builds dept tree structure.
func (s *Service) Tree(ctx context.Context) ([]*TreeNode, error) {
	cols := dao.SysDept.Columns()

	var depts []*entity.SysDept
	err := dao.SysDept.Ctx(ctx).
		WhereNull(cols.DeletedAt).
		Order(cols.OrderNum + " ASC").
		Scan(&depts)
	if err != nil {
		return nil, err
	}

	// Build tree from flat list
	nodeMap := make(map[int]*TreeNode)
	for _, d := range depts {
		nodeMap[d.Id] = &TreeNode{
			Id:       d.Id,
			Label:    d.Name,
			Children: make([]*TreeNode, 0),
		}
	}

	var roots []*TreeNode
	for _, d := range depts {
		node := nodeMap[d.Id]
		if parent, ok := nodeMap[d.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// ExcludeInput defines input for Exclude function.
type ExcludeInput struct {
	Id int // 要排除的部门ID
}

// Exclude returns dept list excluding specified dept and its descendants.
func (s *Service) Exclude(ctx context.Context, in ExcludeInput) ([]*entity.SysDept, error) {
	// Get the target dept
	dept, err := s.GetById(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	cols := dao.SysDept.Columns()
	prefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)

	// Get all non-deleted depts excluding the target and its descendants
	var list []*entity.SysDept
	err = dao.SysDept.Ctx(ctx).
		WhereNull(cols.DeletedAt).
		WhereNot(cols.Id, in.Id).
		WhereNotLike(cols.Ancestors, prefix+",%").
		WhereNotLike(cols.Ancestors, prefix).
		Order(cols.OrderNum + " ASC").
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// Users gets users for leader selection.
// When deptId=0, returns all users. When deptId>0, returns users in the dept and all its sub-depts.
// Supports keyword search on username/nickname and result limit.
func (s *Service) Users(ctx context.Context, deptId int, keyword string, limit int) ([]*DeptUser, error) {
	uCols := dao.SysUser.Columns()

	if deptId == 0 {
		// Return all users (for new dept creation)
		q := dao.SysUser.Ctx(ctx).
			Fields(uCols.Id, uCols.Username, uCols.Nickname).
			WhereNull(uCols.DeletedAt)
		if keyword != "" {
			q = q.Where(
				fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", uCols.Username, uCols.Nickname),
				"%"+keyword+"%", "%"+keyword+"%",
			)
		}
		if limit > 0 {
			q = q.Limit(limit)
		}
		var users []*entity.SysUser
		if err := q.Scan(&users); err != nil {
			return nil, err
		}
		result := make([]*DeptUser, 0, len(users))
		for _, u := range users {
			result = append(result, &DeptUser{
				Id:       u.Id,
				Username: u.Username,
				Nickname: u.Nickname,
			})
		}
		return result, nil
	}

	// Collect the selected dept and all its descendant depts via parent_id (cross-database compatible).
	var (
		deptCols  = dao.SysDept.Columns()
		deptIds   = []int{deptId}
		parentIds = []int{deptId}
	)
	for len(parentIds) > 0 {
		childValues, err := dao.SysDept.Ctx(ctx).
			WhereNull(deptCols.DeletedAt).
			WhereIn(deptCols.ParentId, parentIds).
			Fields(deptCols.Id).
			Array()
		if err != nil {
			return nil, err
		}
		var childIds = gconv.Ints(childValues)
		deptIds = append(deptIds, childIds...)
		parentIds = childIds
	}

	// Query sys_user_dept for user_ids in the subtree
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().DeptId, deptIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	if len(userDepts) == 0 {
		return make([]*DeptUser, 0), nil
	}

	// Deduplicate user IDs
	seen := make(map[int]struct{})
	userIds := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		if _, ok := seen[ud.UserId]; !ok {
			seen[ud.UserId] = struct{}{}
			userIds = append(userIds, ud.UserId)
		}
	}

	// Query sys_user for those IDs
	q := dao.SysUser.Ctx(ctx).
		Fields(uCols.Id, uCols.Username, uCols.Nickname).
		WhereIn(uCols.Id, userIds).
		WhereNull(uCols.DeletedAt)
	if keyword != "" {
		q = q.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", uCols.Username, uCols.Nickname),
			"%"+keyword+"%", "%"+keyword+"%",
		)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var users []*entity.SysUser
	if err := q.Scan(&users); err != nil {
		return nil, err
	}

	// Convert to DeptUser
	result := make([]*DeptUser, 0, len(users))
	for _, u := range users {
		result = append(result, &DeptUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}

	return result, nil
}

// UserDeptTree builds dept tree with user count per node, plus an "未分配部门" virtual node.
func (s *Service) UserDeptTree(ctx context.Context) ([]*TreeNode, error) {
	// Get base tree
	nodes, err := s.Tree(ctx)
	if err != nil {
		return nil, err
	}

	// Get user count per dept via sys_user_dept (only count non-deleted users)
	type DeptCount struct {
		DeptId int `json:"dept_id"`
		Cnt    int `json:"cnt"`
	}
	var counts []DeptCount
	err = dao.SysUserDept.Ctx(ctx).
		Fields("dept_id, COUNT(*) as cnt").
		InnerJoin(
			dao.SysUser.Table(),
			fmt.Sprintf(
				"%s.%s = %s.%s",
				dao.SysUserDept.Table(), dao.SysUserDept.Columns().UserId,
				dao.SysUser.Table(), dao.SysUser.Columns().Id,
			),
		).
		Where(fmt.Sprintf("%s.%s IS NULL", dao.SysUser.Table(), dao.SysUser.Columns().DeletedAt)).
		Group("dept_id").
		Scan(&counts)
	if err != nil {
		return nil, err
	}
	countMap := make(map[int]int)
	for _, c := range counts {
		countMap[c.DeptId] = c.Cnt
	}

	// Apply user counts to tree nodes (parent = self + all descendants)
	var applyCount func(nodes []*TreeNode)
	applyCount = func(nodes []*TreeNode) {
		for _, n := range nodes {
			applyCount(n.Children)
			n.UserCount = countMap[n.Id]
			for _, child := range n.Children {
				n.UserCount += child.UserCount
			}
			n.Label = fmt.Sprintf("%s(%d)", n.Label, n.UserCount)
		}
	}
	applyCount(nodes)

	// Count unassigned users (users not in sys_user_dept)
	uCols := dao.SysUser.Columns()
	totalUsers, err := dao.SysUser.Ctx(ctx).WhereNull(uCols.DeletedAt).Count()
	if err != nil {
		return nil, err
	}
	assignedUsers := 0
	for _, c := range countMap {
		assignedUsers += c
	}
	unassignedCount := totalUsers - assignedUsers

	// Append "未分配部门" virtual node at the end
	unassignedNode := &TreeNode{
		Id:        0,
		Label:     fmt.Sprintf("未分配部门(%d)", unassignedCount),
		UserCount: unassignedCount,
		Children:  make([]*TreeNode, 0),
	}
	result := make([]*TreeNode, 0, len(nodes)+1)
	result = append(result, nodes...)
	result = append(result, unassignedNode)

	return result, nil
}

// checkCodeUnique checks if the dept code is unique (excluding the given dept ID for updates).
func (s *Service) checkCodeUnique(ctx context.Context, code string, excludeId int) error {
	cols := dao.SysDept.Columns()
	m := dao.SysDept.Ctx(ctx).
		Where(cols.Code, code).
		WhereNull(cols.DeletedAt)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("部门编码已存在")
	}
	return nil
}
