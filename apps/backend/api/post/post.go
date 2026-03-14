// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package post

import (
	"context"

	"backend/api/post/v1"
)

type IPostV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error)
	DeptTree(ctx context.Context, req *v1.DeptTreeReq) (res *v1.DeptTreeRes, err error)
	OptionSelect(ctx context.Context, req *v1.OptionSelectReq) (res *v1.OptionSelectRes, err error)
}
