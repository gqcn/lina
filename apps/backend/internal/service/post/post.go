package post

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"backend/internal/dao"
	"backend/internal/model/do"
	"backend/internal/model/entity"
)

// Service provides post management operations.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum  int
	PageSize int
	DeptId   *int
	Code     string
	Name     string
	Status   *int
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysPost
	Total int
}

// List queries post list with pagination and filters.
func (s *Service) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysPost.Columns()
		m    = dao.SysPost.Ctx(ctx).WhereNull(cols.DeletedAt)
	)

	// Apply filters
	if in.DeptId != nil {
		if *in.DeptId == -1 {
			// Unassigned: posts with dept_id = 0
			m = m.Where(cols.DeptId, 0)
		} else {
			m = m.Where(cols.DeptId, *in.DeptId)
		}
	}
	if in.Code != "" {
		m = m.WhereLike(cols.Code, "%"+in.Code+"%")
	}
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysPost
	err = m.Page(in.PageNum, in.PageSize).
		Order(cols.Sort + " ASC").
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
	DeptId int
	Code   string
	Name   string
	Sort   int
	Status int
	Remark string
}

// Create creates a new post.
func (s *Service) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check code uniqueness
	cols := dao.SysPost.Columns()
	count, err := dao.SysPost.Ctx(ctx).
		Where(do.SysPost{Code: in.Code}).
		WhereNull(cols.DeletedAt).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("岗位编码已存在")
	}

	// Insert post
	id, err := dao.SysPost.Ctx(ctx).Data(do.SysPost{
		DeptId:    in.DeptId,
		Code:      in.Code,
		Name:      in.Name,
		Sort:      in.Sort,
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

// GetById retrieves post by ID.
func (s *Service) GetById(ctx context.Context, id int) (*entity.SysPost, error) {
	var post *entity.SysPost
	cols := dao.SysPost.Columns()
	err := dao.SysPost.Ctx(ctx).
		Where(do.SysPost{Id: id}).
		WhereNull(cols.DeletedAt).
		Scan(&post)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, gerror.New("岗位不存在")
	}
	return post, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int
	DeptId *int
	Code   *string
	Name   *string
	Sort   *int
	Status *int
	Remark *string
}

// Update updates post information.
func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	// Check post exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysPost{
		UpdatedAt: gtime.Now(),
	}
	if in.DeptId != nil {
		data.DeptId = *in.DeptId
	}
	if in.Code != nil {
		data.Code = *in.Code
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysPost.Ctx(ctx).Where(do.SysPost{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes posts by comma-separated IDs.
func (s *Service) Delete(ctx context.Context, ids string) error {
	idList := gstr.SplitAndTrim(ids, ",")
	if len(idList) == 0 {
		return gerror.New("请选择要删除的岗位")
	}

	cols := dao.SysUserPost.Columns()
	var validIds []int
	for _, idStr := range idList {
		id := gconv.Int(idStr)
		if id == 0 {
			continue
		}

		// Check if post is assigned to users
		count, err := dao.SysUserPost.Ctx(ctx).
			Where(cols.PostId, id).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.Newf("岗位ID %d 已分配给用户，不能删除", id)
		}
		validIds = append(validIds, id)
	}

	if len(validIds) == 0 {
		return gerror.New("没有有效的岗位ID")
	}

	// Soft delete all valid ids
	postCols := dao.SysPost.Columns()
	_, err := dao.SysPost.Ctx(ctx).
		WhereIn(postCols.Id, validIds).
		Data(do.SysPost{DeletedAt: gtime.Now()}).
		Update()
	return err
}

// DeptTreeNode defines a department tree node.
type DeptTreeNode struct {
	Id       int             `json:"id"`
	Label    string          `json:"label"`
	Children []*DeptTreeNode `json:"children"`
}

// DeptTree returns department tree structure with "未分配部门" virtual node.
func (s *Service) DeptTree(ctx context.Context) ([]*DeptTreeNode, error) {
	cols := dao.SysDept.Columns()
	var depts []*entity.SysDept
	err := dao.SysDept.Ctx(ctx).
		WhereNull(cols.DeletedAt).
		Order(cols.OrderNum + " ASC").
		Scan(&depts)
	if err != nil {
		return nil, err
	}

	// Build tree
	nodeMap := make(map[int]*DeptTreeNode)
	for _, d := range depts {
		nodeMap[d.Id] = &DeptTreeNode{
			Id:       d.Id,
			Label:    d.Name,
			Children: make([]*DeptTreeNode, 0),
		}
	}

	var roots []*DeptTreeNode
	for _, d := range depts {
		node := nodeMap[d.Id]
		if parent, ok := nodeMap[d.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	// Append "未分配部门" virtual node
	unassignedNode := &DeptTreeNode{
		Id:       -1,
		Label:    "未分配部门",
		Children: make([]*DeptTreeNode, 0),
	}
	roots = append(roots, unassignedNode)

	return roots, nil
}

// PostOption defines a post option for select dropdown.
type PostOption struct {
	PostId   int    `json:"postId"`
	PostName string `json:"postName"`
}

// OptionSelectInput defines input for OptionSelect function.
type OptionSelectInput struct {
	DeptId *int
}

// OptionSelect returns post options for select dropdown.
func (s *Service) OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error) {
	cols := dao.SysPost.Columns()
	m := dao.SysPost.Ctx(ctx).
		Where(cols.Status, 1).
		WhereNull(cols.DeletedAt)

	if in.DeptId != nil {
		m = m.Where(cols.DeptId, *in.DeptId)
	}

	var list []*entity.SysPost
	err := m.Order(cols.Sort + " ASC").Scan(&list)
	if err != nil {
		return nil, err
	}

	options := make([]PostOption, 0, len(list))
	for _, p := range list {
		options = append(options, PostOption{
			PostId:   p.Id,
			PostName: p.Name,
		})
	}

	return options, nil
}