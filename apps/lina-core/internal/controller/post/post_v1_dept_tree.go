package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// DeptTree 获取部门树形结构（含岗位数量）
func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.postSvc.DeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{
		List: convertDeptTreeNodes(nodes),
	}, nil
}

// convertDeptTreeNodes 将服务层DeptTreeNode切片转换为API层DeptTreeNode切片
func convertDeptTreeNodes(nodes []*postsvc.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.DeptTreeNode{
			Id:        n.Id,
			Label:     n.Label,
			PostCount: n.PostCount,
			Children:  convertDeptTreeNodes(n.Children),
		})
	}
	return result
}
