// Package dynamicservice implements guest-side backend services for the
// plugin-demo-dynamic sample plugin.
package dynamicservice

import (
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/plugindb"
)

// Service encapsulates the dynamic plugin backend business logic.
type Service struct {
	runtimeSvc *pluginbridge.RuntimeHostService
	storageSvc *pluginbridge.StorageHostService
	httpSvc    *pluginbridge.HTTPHostService
	dataSvc    *plugindb.DB
}

// New creates and returns a new dynamic plugin backend service.
func New() *Service {
	return &Service{
		runtimeSvc: pluginbridge.Runtime(),
		storageSvc: pluginbridge.Storage(),
		httpSvc:    pluginbridge.HTTP(),
		dataSvc:    plugindb.Open(),
	}
}
