package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// Tree 获取部门树形结构
func (c *ControllerV1) Tree(ctx context.Context, req *v1.TreeReq) (res *v1.TreeRes, err error) {
	nodes, err := c.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TreeRes{
		List: convertTreeNodes(nodes),
	}, nil
}

// convertTreeNodes 将服务层TreeNode切片转换为API层TreeNode切片
func convertTreeNodes(nodes []*deptsvc.TreeNode) []*v1.TreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.TreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.TreeNode{
			Id:       n.Id,
			Label:    n.Label,
			Children: convertTreeNodes(n.Children),
		})
	}
	return result
}
