package dynamic

// Service provides review-only backend example services for the dynamic sample plugin.
type Service struct{}

// New creates and returns a new dynamic review service.
func New() *Service {
	return &Service{}
}
