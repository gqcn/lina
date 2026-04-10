package dynamic

import "context"

const reviewSummaryMessage = "This backend example shows how a dynamic plugin can organize Go APIs before phase-3 runtime execution is implemented."

// ReviewSummaryOutput defines one review-only dynamic backend summary.
type ReviewSummaryOutput struct {
	// Message describes the current purpose of the backend sample code.
	Message string
}

// ReviewSummary returns one review-only backend summary payload.
func (s *Service) ReviewSummary(ctx context.Context) (out *ReviewSummaryOutput, err error) {
	return &ReviewSummaryOutput{
		Message: reviewSummaryMessage,
	}, nil
}
