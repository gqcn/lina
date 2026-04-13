// This file executes dynamic bridge requests inside the active Wasm runtime.

package plugin

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

type dynamicWasmExecutor struct {
	service *Service
}

func (e *dynamicWasmExecutor) Execute(
	ctx context.Context,
	manifest *pluginManifest,
	request *pluginbridge.BridgeRequestEnvelopeV1,
) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	_ = ctx
	if manifest == nil || manifest.RuntimeArtifact == nil || manifest.BridgeSpec == nil {
		return nil, gerror.New("动态插件缺少 bridge 运行时信息")
	}
	if !manifest.BridgeSpec.RouteExecution {
		return pluginbridge.NewFailureResponse(501, "BRIDGE_NOT_IMPLEMENTED", "Dynamic route bridge is disabled"), nil
	}

	content, err := pluginbridge.EncodeRequestEnvelope(request)
	if err != nil {
		return nil, err
	}
	return e.service.executeDynamicWasmBridge(ctx, manifest, content)
}
