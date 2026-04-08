package demo

import (
	"context"

	"lina-plugin-demo/backend/api/demo/v1"
)

// IDemoV1 defines plugin-demo demo APIs.
type IDemoV1 interface {
	// Summary returns one concise plugin summary for page rendering and route verification.
	Summary(ctx context.Context, req *v1.SummaryReq) (res *v1.SummaryRes, err error)
}
