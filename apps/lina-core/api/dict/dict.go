// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package dict

import (
	"context"

	"lina-core/api/dict/v1"
)

type IDictV1 interface {
	DataByType(ctx context.Context, req *v1.DataByTypeReq) (res *v1.DataByTypeRes, err error)
	DataCreate(ctx context.Context, req *v1.DataCreateReq) (res *v1.DataCreateRes, err error)
	DataDelete(ctx context.Context, req *v1.DataDeleteReq) (res *v1.DataDeleteRes, err error)
	DataExport(ctx context.Context, req *v1.DataExportReq) (res *v1.DataExportRes, err error)
	DataGet(ctx context.Context, req *v1.DataGetReq) (res *v1.DataGetRes, err error)
	DataList(ctx context.Context, req *v1.DataListReq) (res *v1.DataListRes, err error)
	DataUpdate(ctx context.Context, req *v1.DataUpdateReq) (res *v1.DataUpdateRes, err error)
	TypeCreate(ctx context.Context, req *v1.TypeCreateReq) (res *v1.TypeCreateRes, err error)
	TypeDelete(ctx context.Context, req *v1.TypeDeleteReq) (res *v1.TypeDeleteRes, err error)
	TypeExport(ctx context.Context, req *v1.TypeExportReq) (res *v1.TypeExportRes, err error)
	TypeGet(ctx context.Context, req *v1.TypeGetReq) (res *v1.TypeGetRes, err error)
	TypeList(ctx context.Context, req *v1.TypeListReq) (res *v1.TypeListRes, err error)
	TypeOptions(ctx context.Context, req *v1.TypeOptionsReq) (res *v1.TypeOptionsRes, err error)
	TypeUpdate(ctx context.Context, req *v1.TypeUpdateReq) (res *v1.TypeUpdateRes, err error)
}
