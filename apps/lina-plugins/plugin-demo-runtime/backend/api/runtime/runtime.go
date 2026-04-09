package runtimeapi

import (
	"context"

	"lina-plugin-demo-runtime/backend/api/runtime/v1"
)

// IRuntimeV1 defines the review-only backend example contract for the runtime sample plugin.
type IRuntimeV1 interface {
	ReviewSummary(ctx context.Context, req *v1.ReviewSummaryReq) (res *v1.ReviewSummaryRes, err error)
}
