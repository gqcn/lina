package demo

// Service provides plugin-demo demo services.
type Service struct{}

// New creates and returns a new demo service.
func New() *Service {
	return &Service{}
}
