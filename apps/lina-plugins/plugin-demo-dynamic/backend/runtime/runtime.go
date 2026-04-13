// Package runtime provides the exported guest runtime dispatcher for the
// dynamic sample plugin.
package runtime

import (
	"lina-core/pkg/pluginbridge"
	"lina-plugin-demo-dynamic/backend/internal/controller/dynamic"
)

var ctrl = dynamic.New()

// HandleRequest dispatches a bridge request to the matching dynamic route
// controller.
func HandleRequest(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	if request == nil || request.Route == nil {
		return pluginbridge.NewBadRequestResponse("Dynamic bridge request is missing route metadata"), nil
	}

	switch request.Route.InternalPath {
	case "/backend-summary":
		return ctrl.BackendSummary(request)
	case "/host-call-demo":
		return ctrl.HostCallDemo(request)
	default:
		return pluginbridge.NewNotFoundResponse("Dynamic bridge route not found"), nil
	}
}
