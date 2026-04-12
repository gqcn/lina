package plugin

import (
	pluginapi "lina-core/api/plugin"
	pluginsvc "lina-core/internal/service/plugin"
)

// PublicControllerV1 is the public plugin controller.
type PublicControllerV1 struct {
	pluginSvc *pluginsvc.Service // plugin service
}

// NewPublicV1 creates and returns a new public plugin controller instance.
func NewPublicV1(topologies ...pluginsvc.Topology) pluginapi.IPluginPublicV1 {
	return &PublicControllerV1{
		pluginSvc: pluginsvc.New(topologies...),
	}
}
