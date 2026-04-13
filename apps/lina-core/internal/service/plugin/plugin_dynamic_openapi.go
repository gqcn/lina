// This file projects enabled dynamic plugin routes into the host OpenAPI model.

package plugin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/pkg/pluginbridge"
)

// ProjectDynamicRoutesToOpenAPI projects currently enabled dynamic plugin routes into the host OpenAPI paths.
func (s *Service) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	manifests, err := s.scanPluginManifests()
	if err != nil {
		return err
	}
	if paths == nil {
		return nil
	}

	runtime, err := s.buildFilterRuntimeFromManifests(ctx, manifests)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeDynamic {
			continue
		}
		if !runtime.isEnabled(manifest.ID) {
			continue
		}
		activeManifest, err := s.getActivePluginManifest(ctx, manifest.ID)
		if err != nil || activeManifest == nil {
			continue
		}
		for _, route := range activeManifest.Routes {
			if route == nil {
				continue
			}
			publicPath := buildDynamicRoutePublicPath(activeManifest.ID, route.Path)
			pathItem, ok := paths[publicPath]
			if !ok {
				pathItem = goai.Path{}
			}
			operation := buildDynamicRouteOpenAPIOperation(activeManifest.ID, route, activeManifest.BridgeSpec)
			switch strings.ToUpper(strings.TrimSpace(route.Method)) {
			case http.MethodGet:
				pathItem.Get = operation
			case http.MethodPost:
				pathItem.Post = operation
			case http.MethodPut:
				pathItem.Put = operation
			case http.MethodDelete:
				pathItem.Delete = operation
			}
			paths[publicPath] = pathItem
		}
	}
	return nil
}

func buildDynamicRoutePublicPath(pluginID string, routePath string) string {
	return pluginDynamicRoutePublicPrefix + "/" + strings.TrimSpace(pluginID) + normalizeDynamicRoutePath(routePath)
}

func buildDynamicRouteOpenAPIOperation(
	pluginID string,
	route *pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) *goai.Operation {
	if route == nil {
		return nil
	}
	operation := &goai.Operation{
		Tags:        append([]string(nil), route.Tags...),
		Summary:     route.Summary,
		Description: route.Description,
		OperationID: pluginID + "_" + strings.ToLower(route.Method) + "_" + strings.ReplaceAll(strings.Trim(route.Path, "/"), "/", "_"),
		Responses: goai.Responses{
			"500": goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route execution failed"}},
		},
	}
	if bridgeSpec != nil && bridgeSpec.RouteExecution {
		operation.Responses["200"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route response"}}
	} else {
		operation.Responses["501"] = goai.ResponseRef{Value: &goai.Response{Description: "Dynamic plugin route bridge is not executable"}}
	}
	if route.Access == pluginbridge.AccessLogin {
		operation.Security = &goai.SecurityRequirements{{"BearerAuth": {}}}
	}
	return operation
}
