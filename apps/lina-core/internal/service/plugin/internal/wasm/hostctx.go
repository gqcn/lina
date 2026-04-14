// This file defines the per-request host call context injected into
// context.Context so that wazero host function callbacks can access
// plugin identity and capability permissions.

package wasm

import "context"

// hostCallContextKey is the private context key for host call state.
type hostCallContextKey struct{}

// hostCallContext carries per-request state into wazero host function callbacks.
type hostCallContext struct {
	// pluginID identifies the calling plugin.
	pluginID string
	// capabilities is the set of granted host capabilities for this plugin.
	capabilities map[string]struct{}
}

// withHostCallContext attaches a host call context to the given context.
func withHostCallContext(ctx context.Context, hcc *hostCallContext) context.Context {
	return context.WithValue(ctx, hostCallContextKey{}, hcc)
}

// hostCallContextFrom extracts the host call context from the given context.
func hostCallContextFrom(ctx context.Context) *hostCallContext {
	if hcc, ok := ctx.Value(hostCallContextKey{}).(*hostCallContext); ok {
		return hcc
	}
	return nil
}

// hasCapability checks if the plugin has been granted a specific capability.
func (hcc *hostCallContext) hasCapability(capability string) bool {
	if hcc == nil || hcc.capabilities == nil {
		return false
	}
	_, ok := hcc.capabilities[capability]
	return ok
}
