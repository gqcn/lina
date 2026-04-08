// This file normalizes supported plugin type values and provides lightweight
// helpers used by validation and lifecycle code paths.

package plugin

import "strings"

// pluginTypeValue defines the normalized top-level plugin type enum.
type pluginTypeValue string

const (
	pluginTypeSource      pluginTypeValue = "source"
	pluginTypeRuntime     pluginTypeValue = "runtime"
	pluginRuntimeKindWasm pluginTypeValue = "wasm"
)

// String returns the canonical plugin type value.
func (value pluginTypeValue) String() string {
	return string(value)
}

func normalizePluginType(value string) pluginTypeValue {
	normalizedValue := strings.TrimSpace(strings.ToLower(value))
	switch normalizedValue {
	case pluginRuntimeKindWasm.String(), pluginTypeRuntime.String():
		return pluginTypeRuntime
	case pluginTypeSource.String():
		return pluginTypeSource
	default:
		return pluginTypeValue(normalizedValue)
	}
}

func isSupportedPluginType(value string) bool {
	switch normalizePluginType(value) {
	case pluginTypeSource, pluginTypeRuntime:
		return true
	default:
		return false
	}
}
