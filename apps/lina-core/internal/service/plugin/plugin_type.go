// This file normalizes supported plugin type values and provides lightweight
// helpers used by validation and lifecycle code paths.

package plugin

import "strings"

// pluginTypeValue defines the normalized top-level plugin type enum.
type pluginTypeValue string

const (
	pluginTypeSource        pluginTypeValue = "source"
	pluginTypeDynamic       pluginTypeValue = "dynamic"
	pluginTypeLegacyRuntime pluginTypeValue = "runtime"
	pluginDynamicKindWasm   pluginTypeValue = "wasm"
)

// String returns the canonical plugin type value.
func (value pluginTypeValue) String() string {
	return string(value)
}

func normalizePluginType(value string) pluginTypeValue {
	normalizedValue := strings.TrimSpace(strings.ToLower(value))
	switch normalizedValue {
	case pluginDynamicKindWasm.String(), pluginTypeDynamic.String(), pluginTypeLegacyRuntime.String():
		return pluginTypeDynamic
	case pluginTypeSource.String():
		return pluginTypeSource
	default:
		return pluginTypeValue(normalizedValue)
	}
}

func isSupportedPluginType(value string) bool {
	switch normalizePluginType(value) {
	case pluginTypeSource, pluginTypeDynamic:
		return true
	default:
		return false
	}
}
