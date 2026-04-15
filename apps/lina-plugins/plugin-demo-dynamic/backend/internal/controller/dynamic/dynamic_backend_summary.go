// Backend summary route controller.

package dynamic

import (
	"encoding/json"

	"lina-core/pkg/pluginbridge"
)

// BackendSummary returns plugin bridge execution summary including plugin
// identity, route metadata, and current user context.
func (c *Controller) BackendSummary(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload := c.dynamicSvc.BuildBackendSummaryPayload(request)
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response := pluginbridge.NewJSONResponse(200, content)
	response.Headers = map[string][]string{
		"X-Lina-Plugin-Bridge":     {"plugin-demo-dynamic"},
		"X-Lina-Plugin-Middleware": {"backend-summary"},
	}
	return response, nil
}
