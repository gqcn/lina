package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// IPluginPublicV1 defines public plugin runtime APIs.
type IPluginPublicV1 interface {
	// RuntimeList returns public plugin runtime states for slot rendering.
	RuntimeList(ctx context.Context, req *v1.RuntimeListReq) (res *v1.RuntimeListRes, err error)
}
