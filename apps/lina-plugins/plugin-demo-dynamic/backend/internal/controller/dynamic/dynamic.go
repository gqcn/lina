// Package dynamic implements the dynamic plugin route controllers.

package dynamic

// Controller handles dynamic plugin route requests.
type Controller struct{}

// New creates and returns a new dynamic plugin controller instance.
func New() *Controller {
	return &Controller{}
}
