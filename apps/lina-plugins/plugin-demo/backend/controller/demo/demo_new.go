package demo

import (
	demoapi "lina-plugin-demo/backend/api/demo"
	demosvc "lina-plugin-demo/backend/service/demo"
)

// ControllerV1 is the plugin-demo demo controller.
type ControllerV1 struct {
	demoSvc *demosvc.Service // demo service
}

// NewV1 creates and returns a new plugin-demo demo controller.
func NewV1() demoapi.IDemoV1 {
	return &ControllerV1{
		demoSvc: demosvc.New(),
	}
}
