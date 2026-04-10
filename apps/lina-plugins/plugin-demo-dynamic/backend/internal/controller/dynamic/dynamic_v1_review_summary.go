package dynamic

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// ReviewSummary returns one review-only backend summary payload.
func (c *ControllerV1) ReviewSummary(
	ctx context.Context,
	_ *v1.ReviewSummaryReq,
) (res *v1.ReviewSummaryRes, err error) {
	out, err := c.dynamicSvc.ReviewSummary(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.ReviewSummaryRes{
		Message: out.Message,
	}, nil
}
