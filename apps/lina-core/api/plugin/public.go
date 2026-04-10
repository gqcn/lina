package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// IPluginPublicV1 defines public dynamic-plugin APIs.
type IPluginPublicV1 interface {
	// DynamicList returns public dynamic-plugin states for slot rendering.
	DynamicList(ctx context.Context, req *v1.DynamicListReq) (res *v1.DynamicListRes, err error)
}
