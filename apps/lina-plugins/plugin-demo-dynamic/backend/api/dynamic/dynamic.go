package dynamicapi

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// IDynamicV1 defines the review-only backend example contract for the dynamic sample plugin.
type IDynamicV1 interface {
	ReviewSummary(ctx context.Context, req *v1.ReviewSummaryReq) (res *v1.ReviewSummaryRes, err error)
}
