package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestMatchDynamicRoutePathSupportsParams(t *testing.T) {
	params, ok := matchDynamicRoutePath("/records/{id}/detail", "/records/42/detail")
	if !ok {
		t.Fatal("expected dynamic path match to succeed")
	}
	if params["id"] != "42" {
		t.Fatalf("expected path param id=42, got %#v", params)
	}
}

func TestBuildDynamicRouteOperLogMetadataMapsRouteGovernance(t *testing.T) {
	metadata := buildDynamicRouteOperLogMetadata(&dynamicRouteRuntimeState{
		Match: &dynamicRouteMatch{
			Route: &pluginbridge.RouteContract{
				Tags:    []string{"plugin-review", "dynamic"},
				Summary: "Review summary",
				OperLog: "other",
			},
		},
	})
	if metadata == nil {
		t.Fatal("expected dynamic route operlog metadata to be built")
	}
	if metadata.Title != "plugin-review,dynamic" {
		t.Fatalf("expected title to join route tags, got %q", metadata.Title)
	}
	if metadata.Summary != "Review summary" {
		t.Fatalf("expected summary to be preserved, got %q", metadata.Summary)
	}
	if metadata.OperLogTag != "other" {
		t.Fatalf("expected operlog tag other, got %q", metadata.OperLogTag)
	}
}

func TestParseRuntimeArtifactLoadsRoutesAndBridgeSpec(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-routes",
		"Runtime Route Plugin",
		"v0.3.0",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, buildPluginDynamicArtifactRelativePath("plugin-dynamic-routes"))
	writeRuntimeWasmArtifact(
		t,
		artifactPath,
		&pluginDynamicArtifactManifest{
			ID:      "plugin-dynamic-routes",
			Name:    "Runtime Route Plugin",
			Version: "v0.3.0",
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(defaultTestRuntimeFrontendAssets()),
			RouteCount:         1,
		},
		defaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*pluginbridge.RouteContract{
			{
				Path:        "/review-summary",
				Method:      http.MethodGet,
				Access:      pluginbridge.AccessLogin,
				Permission:  "plugin-dynamic-routes:review:view",
				RequestType: "ReviewSummaryReq",
			},
		},
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact load to succeed, got error: %v", err)
	}
	if len(manifest.Routes) != 1 || manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		t.Fatalf("expected runtime artifact to expose routes and executable bridge, got routes=%d bridge=%#v", len(manifest.Routes), manifest.BridgeSpec)
	}
}

func TestExecuteDynamicWasmBridgeReturnsGuestResponse(t *testing.T) {
	service := New()
	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	artifactPath := filepath.Join(repoRoot, "temp", "output", "plugin-demo-dynamic.wasm")
	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := service.executeDynamicRoute(context.Background(), manifest, &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: "plugin-demo-dynamic",
		Route: &pluginbridge.RouteMatchSnapshotV1{
			InternalPath: "/backend-summary",
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/backend-summary",
			Access:       pluginbridge.AccessLogin,
			Permission:   "plugin-demo-dynamic:backend:view",
		},
		Identity: &pluginbridge.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			IsSuperAdmin: true,
		},
		Request: &pluginbridge.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected dynamic wasm execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected guest bridge response 200, got %#v", response)
	}
	if string(response.Body) == "" {
		t.Fatal("expected guest bridge response body to be non-empty")
	}
	if got := response.Headers["X-Lina-Plugin-Bridge"]; len(got) != 1 || got[0] != "plugin-demo-dynamic" {
		t.Fatalf("expected guest bridge header to be preserved, got %#v", response.Headers)
	}
	if got := response.Headers["X-Lina-Plugin-Middleware"]; len(got) != 1 || got[0] != "backend-summary" {
		t.Fatalf("expected guest-local middleware header to be preserved, got %#v", response.Headers)
	}

	payload := map[string]any{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected guest response body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "plugin-demo-dynamic" {
		t.Fatalf("expected guest payload pluginId to be preserved, got %#v", payload)
	}
	if payload["authenticated"] != true {
		t.Fatalf("expected guest payload authenticated=true, got %#v", payload)
	}
}
