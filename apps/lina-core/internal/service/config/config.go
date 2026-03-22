package config

// Service provides application configuration access.
type Service struct{}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{}
}
