// This file implements the backend summary business logic for the dynamic
// sample plugin.

package dynamicservice

import "lina-core/pkg/pluginbridge"

const backendSummaryMessage = "This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."

// BuildBackendSummaryPayload builds the backend summary response payload.
func (s *Service) BuildBackendSummaryPayload(request *pluginbridge.BridgeRequestEnvelopeV1) map[string]any {
	payload := map[string]any{
		"message":       backendSummaryMessage,
		"pluginId":      request.PluginID,
		"publicPath":    request.Route.PublicPath,
		"access":        request.Route.Access,
		"permission":    request.Route.Permission,
		"authenticated": request.Identity != nil && request.Identity.UserID > 0,
	}
	if request.Identity != nil {
		payload["username"] = request.Identity.Username
		payload["isSuperAdmin"] = request.Identity.IsSuperAdmin
	}
	return payload
}
