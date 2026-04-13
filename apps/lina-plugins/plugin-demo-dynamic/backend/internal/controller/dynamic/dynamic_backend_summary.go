// Backend summary route controller.

package dynamic

import (
	"encoding/json"

	"lina-core/pkg/pluginbridge"
)

const backendSummaryMessage = "This backend example is executed through the plugin-demo-dynamic Wasm bridge runtime."

// BackendSummary returns plugin bridge execution summary including plugin
// identity, route metadata, and current user context.
func (c *Controller) BackendSummary(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
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
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response := pluginbridge.NewJSONResponse(200, content)
	response.Headers = map[string][]string{
		"X-Lina-Plugin-Bridge":    {"plugin-demo-dynamic"},
		"X-Lina-Plugin-Middleware": {"backend-summary"},
	}
	return response, nil
}
