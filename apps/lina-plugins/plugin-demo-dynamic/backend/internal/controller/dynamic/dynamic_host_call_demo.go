// Host call demo route controller.

package dynamic

import (
	"encoding/json"
	"fmt"
	"strconv"

	"lina-core/pkg/pluginbridge"
)

// HostCallDemo demonstrates host function capabilities including structured
// logging via HostLog and persistent state via HostStateGet/HostStateSet.
func (c *Controller) HostCallDemo(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	// Log through host logger.
	username := "anonymous"
	if request.Identity != nil && request.Identity.Username != "" {
		username = request.Identity.Username
	}
	_ = pluginbridge.HostLog(int(pluginbridge.LogLevelInfo), "host call demo invoked", map[string]string{
		"username": username,
	})

	// Read and increment visit counter via host state store.
	visitCount := 0
	current, found, err := pluginbridge.HostStateGet("visit_count")
	if err == nil && found {
		visitCount, _ = strconv.Atoi(current)
	}
	visitCount++
	_ = pluginbridge.HostStateSet("visit_count", fmt.Sprintf("%d", visitCount))

	payload := map[string]any{
		"visitCount": visitCount,
		"pluginId":   request.PluginID,
		"message":    "Host call demo: log written, state incremented.",
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return pluginbridge.NewJSONResponse(200, content), nil
}
