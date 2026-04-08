package plugin

import "strings"

const (
	pluginTypeSource      = "source"
	pluginTypeRuntime     = "runtime"
	pluginRuntimeKindWasm = "wasm"
)

func normalizePluginType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case pluginRuntimeKindWasm, pluginTypeRuntime:
		return pluginTypeRuntime
	case pluginTypeSource:
		return pluginTypeSource
	default:
		return strings.TrimSpace(strings.ToLower(value))
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
