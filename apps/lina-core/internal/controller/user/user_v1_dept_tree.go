package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	deptsvc "lina-core/internal/service/dept"
)

// DeptTree 获取用户部门树形结构
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	svc := deptsvc.New()
	nodes, err := svc.UserDeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{List: convertDeptTreeNodes(nodes)}, nil
}

// convertDeptTreeNodes 将服务层TreeNode切片转换为API层DeptTreeNode切片
func convertDeptTreeNodes(nodes []*deptsvc.TreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.DeptTreeNode{
			Id:        n.Id,
			Label:     n.Label,
			UserCount: n.UserCount,
			Children:  convertDeptTreeNodes(n.Children),
		})
	}
	return result
}
