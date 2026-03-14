package post

import (
	"context"

	v1 "backend/api/post/v1"
	postsvc "backend/internal/service/post"
)

func (c *ControllerV1) DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error) {
	nodes, err := c.postSvc.DeptTree(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeptTreeRes{
		List: convertDeptTreeNodes(nodes),
	}, nil
}

func convertDeptTreeNodes(nodes []*postsvc.DeptTreeNode) []*v1.DeptTreeNode {
	if nodes == nil {
		return nil
	}
	result := make([]*v1.DeptTreeNode, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &v1.DeptTreeNode{
			Id:       n.Id,
			Label:    n.Label,
			Children: convertDeptTreeNodes(n.Children),
		})
	}
	return result
}
