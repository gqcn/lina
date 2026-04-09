package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/model/entity"
)

func TestPluginDemoRuntimePluginMatchesReviewSource(t *testing.T) {
	service := New()
	resetRuntimeFrontendBundleCache()
	t.Cleanup(resetRuntimeFrontendBundleCache)

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-runtime")
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")

	manifest := &pluginManifest{}
	if err = service.loadPluginYAMLFile(manifestPath, manifest); err != nil {
		t.Fatalf("failed to load runtime plugin manifest: %v", err)
	}

	expectedFrontendAssets := mustCollectRuntimeBuilderFrontendAssets(t, service, pluginDir)
	expectedInstallSQLAssets := mustCollectRuntimeBuilderSQLAssets(t, service, pluginDir, false)
	expectedUninstallSQLAssets := mustCollectRuntimeBuilderSQLAssets(t, service, pluginDir, true)
	expectedMetadata := buildExpectedRuntimeReviewMetadata(
		expectedFrontendAssets,
		expectedInstallSQLAssets,
		expectedUninstallSQLAssets,
	)

	stagedPluginDir := stageRuntimePluginForValidation(t, service, pluginDir)
	stagedManifestPath := filepath.Join(stagedPluginDir, "plugin.yaml")
	stagedManifest := &pluginManifest{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		Description:  manifest.Description,
		Author:       manifest.Author,
		Homepage:     manifest.Homepage,
		License:      manifest.License,
		ManifestPath: stagedManifestPath,
		RootDir:      stagedPluginDir,
	}
	if err = service.validatePluginManifest(stagedManifest, stagedManifestPath); err != nil {
		t.Fatalf("expected runtime sample plugin manifest to be valid, got error: %v", err)
	}

	if stagedManifest.RuntimeArtifact == nil || stagedManifest.RuntimeArtifact.Manifest == nil {
		t.Fatalf("expected runtime artifact metadata to be loaded")
	}
	if stagedManifest.RuntimeArtifact.Manifest.ID != "plugin-demo-runtime" {
		t.Fatalf("expected runtime plugin id plugin-demo-runtime, got %s", stagedManifest.RuntimeArtifact.Manifest.ID)
	}
	if stagedManifest.RuntimeArtifact.Manifest.Name != manifest.Name {
		t.Fatalf("expected runtime plugin name %s, got %s", manifest.Name, stagedManifest.RuntimeArtifact.Manifest.Name)
	}
	if stagedManifest.RuntimeArtifact.Manifest.Version != manifest.Version {
		t.Fatalf("expected runtime plugin version %s, got %s", manifest.Version, stagedManifest.RuntimeArtifact.Manifest.Version)
	}
	if stagedManifest.RuntimeArtifact.RuntimeKind != expectedMetadata.RuntimeKind {
		t.Fatalf("expected runtime kind %s, got %s", expectedMetadata.RuntimeKind, stagedManifest.RuntimeArtifact.RuntimeKind)
	}
	if stagedManifest.RuntimeArtifact.ABIVersion != expectedMetadata.ABIVersion {
		t.Fatalf("expected runtime ABI %s, got %s", expectedMetadata.ABIVersion, stagedManifest.RuntimeArtifact.ABIVersion)
	}

	assertRuntimeFrontendAssetsMatch(t, expectedFrontendAssets, stagedManifest.RuntimeArtifact.FrontendAssets)
	assertRuntimeSQLAssetsMatch(t, expectedInstallSQLAssets, stagedManifest.RuntimeArtifact.InstallSQLAssets)
	assertRuntimeSQLAssetsMatch(t, expectedUninstallSQLAssets, stagedManifest.RuntimeArtifact.UninstallSQLAssets)

	bundle, err := service.ensureRuntimeFrontendBundle(context.Background(), stagedManifest)
	if err != nil {
		t.Fatalf("expected runtime frontend bundle to load, got error: %v", err)
	}
	if bundle == nil || !bundle.HasAsset("mount.js") {
		t.Fatalf("expected runtime frontend bundle to expose mount.js")
	}
	if _, statErr := os.Stat(filepath.Join(stagedPluginDir, "runtime", "frontend-assets")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no extracted runtime frontend-assets directory, got err=%v", statErr)
	}

	hostedBaseURL := service.BuildRuntimeFrontendPublicBaseURL(stagedManifest.ID, stagedManifest.Version)
	menus := []*entity.SysMenu{
		{
			MenuKey:    "plugin:plugin-demo-runtime:main-entry",
			Name:       "运行时插件示例",
			Path:       hostedBaseURL + "mount.js",
			Component:  pluginRuntimePageComponentPath,
			QueryParam: `{"pluginAccessMode":"embedded-mount"}`,
			IsFrame:    0,
		},
	}
	if err = service.validateRuntimeHostedMenuBindings(context.Background(), stagedManifest, menus); err != nil {
		t.Fatalf("expected runtime sample menu contract to be valid, got error: %v", err)
	}
}

func stageRuntimePluginForValidation(t *testing.T, service *Service, sourcePluginDir string) string {
	t.Helper()

	buildOut, err := service.BuildRuntimeWasmArtifactFromSource(sourcePluginDir)
	if err != nil {
		t.Fatalf("failed to build source runtime artifact: %v", err)
	}

	targetPluginDir := t.TempDir()
	if err = os.MkdirAll(filepath.Join(targetPluginDir, "runtime"), 0o755); err != nil {
		t.Fatalf("failed to create staged runtime dir: %v", err)
	}

	manifestContent, err := os.ReadFile(filepath.Join(sourcePluginDir, "plugin.yaml"))
	if err != nil {
		t.Fatalf("failed to read source plugin manifest: %v", err)
	}
	if err = os.WriteFile(filepath.Join(targetPluginDir, "plugin.yaml"), manifestContent, 0o644); err != nil {
		t.Fatalf("failed to write staged plugin manifest: %v", err)
	}

	artifactPath := filepath.Join(
		targetPluginDir,
		buildPluginRuntimeArtifactRelativePath(buildOut.Manifest.ID),
	)
	if err = os.WriteFile(artifactPath, buildOut.Content, 0o644); err != nil {
		t.Fatalf("failed to write staged runtime artifact: %v", err)
	}

	return targetPluginDir
}

func buildExpectedRuntimeReviewMetadata(
	frontendAssets []*pluginRuntimeArtifactFrontendAsset,
	installSQLAssets []*pluginRuntimeArtifactSQLAsset,
	uninstallSQLAssets []*pluginRuntimeArtifactSQLAsset,
) *pluginRuntimeArtifactMetadata {
	return &pluginRuntimeArtifactMetadata{
		RuntimeKind:        pluginRuntimeKindWasm.String(),
		ABIVersion:         pluginRuntimeSupportedABIVersion,
		FrontendAssetCount: len(frontendAssets),
		SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
	}
}

func mustCollectRuntimeBuilderFrontendAssets(
	t *testing.T,
	service *Service,
	pluginDir string,
) []*pluginRuntimeArtifactFrontendAsset {
	t.Helper()

	assets, err := service.collectRuntimeBuilderFrontendAssets(pluginDir)
	if err != nil {
		t.Fatalf("failed to collect runtime builder frontend assets: %v", err)
	}
	return assets
}

func mustCollectRuntimeBuilderSQLAssets(
	t *testing.T,
	service *Service,
	pluginDir string,
	uninstall bool,
) []*pluginRuntimeArtifactSQLAsset {
	t.Helper()

	assets, err := service.collectRuntimeBuilderSQLAssets(pluginDir, uninstall)
	if err != nil {
		t.Fatalf("failed to collect runtime builder SQL assets: %v", err)
	}
	return assets
}

func assertRuntimeFrontendAssetsMatch(
	t *testing.T,
	expected []*pluginRuntimeArtifactFrontendAsset,
	actual []*pluginRuntimeArtifactFrontendAsset,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d frontend assets, got %d", len(expected), len(actual))
	}

	expectedByPath := make(map[string]*pluginRuntimeArtifactFrontendAsset, len(expected))
	for _, asset := range expected {
		expectedByPath[asset.Path] = asset
	}

	for _, asset := range actual {
		expectedAsset, ok := expectedByPath[asset.Path]
		if !ok {
			t.Fatalf("unexpected frontend asset path: %s", asset.Path)
		}
		if string(asset.Content) != string(expectedAsset.Content) {
			t.Fatalf("unexpected content for frontend asset %s", asset.Path)
		}
	}
}

func assertRuntimeSQLAssetsMatch(
	t *testing.T,
	expected []*pluginRuntimeArtifactSQLAsset,
	actual []*pluginRuntimeArtifactSQLAsset,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d SQL assets, got %d", len(expected), len(actual))
	}

	for index := range expected {
		if actual[index].Key != expected[index].Key {
			t.Fatalf("expected SQL asset key %s, got %s", expected[index].Key, actual[index].Key)
		}
		if strings.TrimSpace(actual[index].Content) != strings.TrimSpace(expected[index].Content) {
			t.Fatalf("unexpected SQL content for asset %s", expected[index].Key)
		}
	}
}
