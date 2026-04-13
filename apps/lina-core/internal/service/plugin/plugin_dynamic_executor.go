// This file defines dynamic route executors and runtime selection for active
// dynamic plugin releases.

package plugin

import (
	"context"
	"net/http"

	"lina-core/pkg/pluginbridge"
)

// dynamicRouteExecutor executes one encoded bridge request against one active runtime.
type dynamicRouteExecutor interface {
	Execute(ctx context.Context, manifest *pluginManifest, request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error)
}

type dynamicPlaceholderExecutor struct{}

func (e *dynamicPlaceholderExecutor) Execute(
	ctx context.Context,
	manifest *pluginManifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_ = ctx
	_ = manifest
	_ = request
	return pluginbridge.NewFailureResponse(
		http.StatusNotImplemented,
		"BRIDGE_NOT_IMPLEMENTED",
		"Dynamic route bridge is not executable for the active plugin release",
	), nil
}

func (s *Service) executeDynamicRoute(
	ctx context.Context,
	manifest *pluginManifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	executor := s.selectDynamicRouteExecutor(manifest)
	return executor.Execute(ctx, manifest, request)
}

func (s *Service) selectDynamicRouteExecutor(manifest *pluginManifest) dynamicRouteExecutor {
	if manifest == nil || manifest.BridgeSpec == nil {
		return &dynamicPlaceholderExecutor{}
	}
	if manifest.BridgeSpec.RouteExecution && manifest.BridgeSpec.RuntimeKind == pluginbridge.RuntimeKindWasm {
		return &dynamicWasmExecutor{service: s}
	}
	return &dynamicPlaceholderExecutor{}
}
