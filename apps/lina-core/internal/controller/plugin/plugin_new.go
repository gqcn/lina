package plugin

import (
	pluginapi "lina-core/api/plugin"
	pluginsvc "lina-core/internal/service/plugin"
)

// ControllerV1 is the plugin controller.
type ControllerV1 struct {
	pluginSvc *pluginsvc.Service // plugin service
}

// NewV1 creates and returns a new plugin controller instance.
func NewV1(topologies ...pluginsvc.Topology) pluginapi.IPluginV1 {
	return &ControllerV1{
		pluginSvc: pluginsvc.New(topologies...),
	}
}
