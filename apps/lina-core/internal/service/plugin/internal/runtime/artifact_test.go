// This file covers runtime artifact discovery, validation, and parsing behaviors.

package runtime_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

func TestScanPluginManifestsDiscoversRuntimePluginFromStorage(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dynamic-storage-scan"
	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Storage Scan Plugin",
		"v0.9.1",
		nil,
		nil,
	)

	manifests, err := services.Catalog.ScanManifests()
	if err != nil {
		t.Fatalf("expected scan to discover dynamic artifact from storage path, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest == nil || manifest.ID != pluginID {
			continue
		}
		if manifest.RuntimeArtifact == nil {
			t.Fatalf("expected dynamic artifact metadata to be loaded for %s", pluginID)
		}
		return
	}
	t.Fatalf("expected dynamic plugin %s to be discovered from storage path", pluginID)
}

func TestScanPluginManifestsDropsRuntimePluginAfterArtifactRemoval(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dynamic-missing-scan"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Scan Plugin",
		"v0.9.2",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated dynamic artifact: %v", err)
	}

	manifests, err := services.Catalog.ScanManifests()
	if err != nil {
		t.Fatalf("expected scan to succeed after dynamic artifact removal, got error: %v", err)
	}

	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			t.Fatalf("expected removed dynamic plugin %s to disappear from scan results", pluginID)
		}
	}
}

func TestEnsureRuntimeArtifactAvailableRejectsMissingGeneratedWasm(t *testing.T) {
	services := testutil.NewServices()

	pluginID := "plugin-dynamic-missing-install"
	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Missing Install Plugin",
		"v0.9.3",
		nil,
		nil,
	)

	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove generated dynamic artifact: %v", err)
	}

	manifest := &catalog.Manifest{
		ID:      pluginID,
		Name:    "Runtime Missing Install Plugin",
		Version: "v0.9.3",
		Type:    catalog.TypeDynamic.String(),
		RootDir: filepath.Dir(artifactPath),
	}

	strictErr := services.Runtime.ValidateRuntimeArtifact(manifest, filepath.Dir(artifactPath))
	if strictErr == nil || !runtime.IsMissingArtifactError(strictErr) {
		t.Fatalf("expected strict runtime validation to report a missing artifact, got: %v", strictErr)
	}

	err := services.Runtime.EnsureRuntimeArtifactAvailable(manifest, "安装")
	if err == nil {
		t.Fatalf("expected lifecycle guard to reject missing dynamic artifact")
	}
	if expected := "make wasm p=" + pluginID; !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle guard error to mention %q, got: %v", expected, err)
	}
	if expected := filepath.ToSlash(runtime.BuildArtifactRelativePath(pluginID)); !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected lifecycle guard error to mention missing wasm path %q, got: %v", expected, err)
	}
}

func TestParseRuntimeArtifactLoadsRoutesAndBridgeSpec(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := testutil.CreateTestRuntimePluginDir(
		t,
		"plugin-dynamic-routes",
		"Runtime Route Plugin",
		"v0.3.0",
		nil,
		nil,
	)

	artifactPath := filepath.Join(pluginDir, runtime.BuildArtifactRelativePath("plugin-dynamic-routes"))
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      "plugin-dynamic-routes",
			Name:    "Runtime Route Plugin",
			Version: "v0.3.0",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
			RouteCount:         1,
			Capabilities:       []string{pluginbridge.CapabilityRuntime},
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{
						pluginbridge.HostServiceMethodRuntimeLogWrite,
						pluginbridge.HostServiceMethodRuntimeStateGet,
					},
				},
			},
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*pluginbridge.RouteContract{
			{
				Path:        "/review-summary",
				Method:      "GET",
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

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact load to succeed, got error: %v", err)
	}
	if len(manifest.Routes) != 1 || manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		t.Fatalf("expected runtime artifact to expose routes and executable bridge, got routes=%d bridge=%#v", len(manifest.Routes), manifest.BridgeSpec)
	}
	if _, ok := manifest.HostCapabilities[pluginbridge.CapabilityRuntime]; !ok {
		t.Fatalf("expected runtime capability to be restored, got %#v", manifest.HostCapabilities)
	}
	if len(manifest.HostServices) != 1 || manifest.HostServices[0].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("expected runtime host service snapshot to be restored, got %#v", manifest.HostServices)
	}
}
