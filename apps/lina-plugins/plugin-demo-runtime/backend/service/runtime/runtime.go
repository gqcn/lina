package runtime

// Service provides review-only backend example services for the runtime sample plugin.
type Service struct{}

// New creates and returns a new runtime review service.
func New() *Service {
	return &Service{}
}
