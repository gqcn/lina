package plugin

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestEnsureRuntimeFrontendBundleReadsEmbeddedAssetsWithoutExtraction(t *testing.T) {
	service := New()
	resetRuntimeFrontendBundleCache()
	t.Cleanup(resetRuntimeFrontendBundleCache)

	pluginDir := createTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-bundle",
		"Runtime Bundle Plugin",
		"v0.4.0",
		[]*pluginDynamicArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>bundle asset</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('bundle asset');")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifest := &pluginManifest{
		ID:           "plugin-dynamic-bundle",
		Name:         "Runtime Bundle Plugin",
		Version:      "v0.4.0",
		Type:         pluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := service.validatePluginManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	bundle, err := service.ensureRuntimeFrontendBundle(context.Background(), manifest)
	if err != nil {
		t.Fatalf("expected dynamic frontend bundle to load, got error: %v", err)
	}

	indexContent, contentType, err := bundle.ReadAsset("")
	if err != nil {
		t.Fatalf("expected bundle root asset to resolve, got error: %v", err)
	}
	if expected := "<html><body>bundle asset</body></html>"; !strings.Contains(string(indexContent), expected) {
		t.Fatalf("expected bundle index content to contain %q, got %q", expected, string(indexContent))
	}
	if contentType != "text/html; charset=utf-8" {
		t.Fatalf("expected html content type, got %s", contentType)
	}

	assetDir := filepath.Join(pluginDir, "runtime", "frontend-assets")
	if _, statErr := os.Stat(assetDir); !os.IsNotExist(statErr) {
		t.Fatalf("expected no extracted frontend-assets directory, got err=%v", statErr)
	}
}

func TestUpdateStatusEnablesBackendOnlyDynamicPluginWithoutFrontendAssets(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-backend-only"
	)

	resetRuntimeFrontendBundleCache()
	t.Cleanup(resetRuntimeFrontendBundleCache)

	artifactPath := createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Backend Only Dynamic Plugin",
		"v0.4.1",
		nil,
		nil,
		nil,
	)

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		_ = os.Remove(artifactPath)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected backend-only artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected backend-only artifact sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		t.Fatalf("expected backend-only plugin install state to be set, got error: %v", err)
	}

	if err = service.UpdateStatus(ctx, pluginID, pluginStatusEnabled); err != nil {
		t.Fatalf("expected backend-only dynamic plugin enable to succeed, got error: %v", err)
	}
	if !service.IsEnabled(ctx, pluginID) {
		t.Fatalf("expected backend-only dynamic plugin to be enabled after status update")
	}
}

func createTestRuntimeStorageArtifactWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) string {
	return createTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
}

func createTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) string {
	t.Helper()

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	storageDir := filepath.Join(repoRoot, "temp", "output")
	if err = os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, buildPluginDynamicArtifactFileName(pluginID))
	t.Cleanup(func() {
		_ = os.Remove(artifactPath)
	})

	writeRuntimeWasmArtifact(
		t,
		artifactPath,
		&pluginDynamicArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		routeContracts,
		bridgeSpec,
	)
	return artifactPath
}
