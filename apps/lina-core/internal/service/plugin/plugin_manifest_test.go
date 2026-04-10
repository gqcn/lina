// This file contains unit tests for manifest validation, convention-based
// resource discovery, and review-oriented plugin metadata helpers.

package plugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/pluginhost"
)

func TestValidatePluginManifestAcceptsMinimalSourcePlugin(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-manifest-valid")

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{
		ID:          "plugin-manifest-valid",
		Name:        "Manifest Validation Plugin",
		Version:     "0.1.0",
		Type:        pluginTypeSource.String(),
		Description: "A valid source plugin manifest used by unit tests.",
		Author:      "test-suite",
		License:     "Apache-2.0",
	}

	if err := service.validatePluginManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected manifest to be valid, got error: %v", err)
	}
}

func TestValidatePluginManifestRejectsMissingBackendEntryForSourcePlugin(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-missing-backend")
	if err := os.Remove(filepath.Join(pluginDir, "backend", "plugin.go")); err != nil {
		t.Fatalf("failed to remove backend entry: %v", err)
	}

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{
		ID:      "plugin-missing-backend",
		Name:    "Missing Backend Plugin",
		Version: "0.1.0",
		Type:    pluginTypeSource.String(),
	}

	err := service.validatePluginManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "backend/plugin.go") {
		t.Fatalf("expected missing backend entry error, got: %v", err)
	}
}

func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedWasmMetadata(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-valid",
		"Runtime Validation Plugin",
		"v0.2.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-valid.sql", Content: "SELECT 1;"},
		},
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-valid.sql", Content: "SELECT 2;"},
		},
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{
		ID:          "plugin-dynamic-valid",
		Name:        "Runtime Validation Plugin",
		Version:     "v0.2.0",
		Type:        pluginTypeDynamic.String(),
		Description: "A valid dynamic plugin manifest used by unit tests.",
	}

	if err := service.validatePluginManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}
	if manifest.RuntimeArtifact == nil {
		t.Fatalf("expected dynamic artifact metadata to be loaded")
	}
	if manifest.RuntimeArtifact.RuntimeKind != pluginDynamicKindWasm.String() {
		t.Fatalf("expected runtime kind wasm, got %s", manifest.RuntimeArtifact.RuntimeKind)
	}
	if manifest.RuntimeArtifact.ABIVersion != pluginDynamicSupportedABIVersion {
		t.Fatalf("expected ABI version %s, got %s", pluginDynamicSupportedABIVersion, manifest.RuntimeArtifact.ABIVersion)
	}
}

func TestValidatePluginManifestAcceptsRuntimePluginWithEmbeddedFrontendAssets(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDirWithFrontendAssets(
		t,
		"plugin-dynamic-frontend",
		"Runtime Frontend Plugin",
		"v0.2.1",
		[]*pluginDynamicArtifactFrontendAsset{
			{
				Path:          "index.html",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
				ContentType:   "text/html; charset=utf-8",
			},
			{
				Path:          "assets/app.js",
				ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend')")),
				ContentType:   "application/javascript",
			},
		},
		nil,
		nil,
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{
		ID:      "plugin-dynamic-frontend",
		Name:    "Runtime Frontend Plugin",
		Version: "v0.2.1",
		Type:    pluginTypeDynamic.String(),
	}

	if err := service.validatePluginManifest(manifest, manifestFile); err != nil {
		t.Fatalf("expected dynamic frontend manifest to be valid, got error: %v", err)
	}
	if manifest.RuntimeArtifact == nil {
		t.Fatalf("expected dynamic artifact metadata to be loaded")
	}
	if len(manifest.RuntimeArtifact.FrontendAssets) != 2 {
		t.Fatalf("expected 2 frontend assets, got %d", len(manifest.RuntimeArtifact.FrontendAssets))
	}
	if manifest.RuntimeArtifact.FrontendAssets[0].Path != "index.html" {
		t.Fatalf("expected normalized frontend asset path index.html, got %s", manifest.RuntimeArtifact.FrontendAssets[0].Path)
	}
}

func TestValidatePluginManifestRejectsMismatchedRuntimeWasmManifest(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-mismatch",
		"Runtime Mismatch Plugin",
		"v0.3.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	writeRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, buildPluginDynamicArtifactRelativePath("plugin-dynamic-mismatch")),
		&pluginDynamicArtifactManifest{
			ID:      "plugin-dynamic-other",
			Name:    "Runtime Mismatch Plugin",
			Version: "v0.3.0",
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:   pluginDynamicKindWasm.String(),
			ABIVersion:    pluginDynamicSupportedABIVersion,
			SQLAssetCount: 1,
		},
		nil,
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-mismatch.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifestFile := filepath.Join(pluginDir, "plugin.yaml")
	manifest := &pluginManifest{
		ID:      "plugin-dynamic-mismatch",
		Name:    "Runtime Mismatch Plugin",
		Version: "v0.3.0",
		Type:    pluginTypeDynamic.String(),
	}

	err := service.validatePluginManifest(manifest, manifestFile)
	if err == nil || !strings.Contains(err.Error(), "嵌入清单 ID") {
		t.Fatalf("expected embedded manifest mismatch error, got: %v", err)
	}
}

func TestScanPluginManifestsRejectsDuplicatePluginIDs(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-duplicate-id")

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	manifestContent := strings.Join([]string{
		"id: plugin-demo",
		"name: Duplicate Plugin",
		"version: 0.1.0",
		"type: source",
		"description: Duplicate id test plugin",
		"author: test-suite",
		"license: Apache-2.0",
		"",
	}, "\n")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("failed to write duplicate manifest: %v", err)
	}

	_, err := service.scanPluginManifests()
	if err == nil || !strings.Contains(err.Error(), "插件ID重复") {
		t.Fatalf("expected duplicate plugin id error, got: %v", err)
	}
}

func TestScanPluginManifestsRejectsDuplicateRuntimeArtifactPluginIDs(t *testing.T) {
	service := New()

	createTestRuntimeStorageArtifactWithFilename(
		t,
		"plugin-dynamic-duplicate-a.wasm",
		"plugin-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)
	createTestRuntimeStorageArtifactWithFilename(
		t,
		"plugin-dynamic-duplicate-b.wasm",
		"plugin-dynamic-duplicate",
		"Runtime Duplicate Plugin",
		"v0.1.0",
		nil,
		nil,
	)

	_, err := service.scanPluginManifests()
	if err == nil || !strings.Contains(err.Error(), "动态插件ID重复") {
		t.Fatalf("expected duplicate dynamic plugin id error, got: %v", err)
	}
}

func TestStoreUploadedRuntimePackageWritesCanonicalWasmIntoRuntimeStorage(t *testing.T) {
	service := New()
	ctx := context.Background()

	pluginID := "plugin-dynamic-upload-storage"
	content := buildTestRuntimeWasmArtifactContent(
		t,
		&pluginDynamicArtifactManifest{
			ID:      pluginID,
			Name:    "Runtime Upload Storage Plugin",
			Version: "v0.5.0",
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginDynamicSupportedABIVersion,
			FrontendAssetCount: len(defaultTestRuntimeFrontendAssets()),
		},
		defaultTestRuntimeFrontendAssets(),
		nil,
		nil,
	)

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	storageArtifactPath := filepath.Join(repoRoot, "temp", "output", buildPluginDynamicArtifactFileName(pluginID))
	_ = os.Remove(storageArtifactPath)
	t.Cleanup(func() {
		_ = os.Remove(storageArtifactPath)
	})
	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	out, err := service.storeUploadedRuntimePackage(ctx, "blob", content, true)
	if err != nil {
		t.Fatalf("expected runtime upload to succeed, got error: %v", err)
	}
	if out.Id != pluginID {
		t.Fatalf("expected uploaded plugin id %s, got %s", pluginID, out.Id)
	}
	if !gfile.Exists(storageArtifactPath) {
		t.Fatalf("expected dynamic artifact to be written into storage path: %s", storageArtifactPath)
	}
	if sourceManifestPath := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID, "plugin.yaml"); gfile.Exists(sourceManifestPath) {
		t.Fatalf("expected upload to stop creating source-tree plugin manifests, found: %s", sourceManifestPath)
	}
}

func TestDiscoverPluginSQLPathsUsesDirectoryConvention(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-sql-convention")

	installPaths := service.discoverPluginSQLPaths(pluginDir, false)
	if len(installPaths) != 1 || installPaths[0] != "manifest/sql/001-plugin-sql-convention.sql" {
		t.Fatalf("unexpected install sql paths: %#v", installPaths)
	}

	uninstallPaths := service.discoverPluginSQLPaths(pluginDir, true)
	if len(uninstallPaths) != 1 || uninstallPaths[0] != "manifest/sql/uninstall/001-plugin-sql-convention.sql" {
		t.Fatalf("unexpected uninstall sql paths: %#v", uninstallPaths)
	}
}

func TestDiscoverPluginVuePathsUseDirectoryConvention(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-vue-convention")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	writeTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	pagePaths := service.discoverPluginPagePaths(pluginDir)
	if len(pagePaths) != 1 || pagePaths[0] != "frontend/pages/main-entry.vue" {
		t.Fatalf("unexpected page paths: %#v", pagePaths)
	}

	slotPaths := service.discoverPluginSlotPaths(pluginDir)
	if len(slotPaths) != 1 || slotPaths[0] != "frontend/slots/dashboard.workspace.after/workspace-card.vue" {
		t.Fatalf("unexpected slot paths: %#v", slotPaths)
	}
}

func TestBuildPluginManifestSnapshotIncludesDirectoryDiscoveredAssets(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-snapshot")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	writeTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	snapshot, err := service.buildPluginManifestSnapshot(&pluginManifest{
		ID:           "plugin-snapshot",
		Name:         "Snapshot Plugin",
		Version:      "0.1.0",
		Type:         pluginTypeSource.String(),
		Description:  "Snapshot test plugin",
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	})
	if err != nil {
		t.Fatalf("expected snapshot to build, got error: %v", err)
	}

	for _, expected := range []string{
		"frontendPageCount: 1",
		"frontendSlotCount: 1",
		"installSqlCount: 1",
	} {
		if !strings.Contains(snapshot, expected) {
			t.Fatalf("expected snapshot to contain %s, got: %s", expected, snapshot)
		}
	}
}

func TestBuildPluginManifestSnapshotIncludesRuntimeArtifactMetadata(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-snapshot",
		"Runtime Snapshot Plugin",
		"v0.4.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-snapshot.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &pluginManifest{
		ID:           "plugin-dynamic-snapshot",
		Name:         "Runtime Snapshot Plugin",
		Version:      "v0.4.0",
		Type:         pluginTypeDynamic.String(),
		Description:  "Runtime snapshot test plugin",
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := service.validateRuntimePluginArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	snapshot, err := service.buildPluginManifestSnapshot(manifest)
	if err != nil {
		t.Fatalf("expected snapshot to build, got error: %v", err)
	}

	for _, expected := range []string{
		"runtimeKind: wasm",
		"runtimeAbiVersion: v1",
		"runtimeFrontendAssetCount: 2",
		"runtimeSqlAssetCount: 1",
	} {
		if !strings.Contains(snapshot, expected) {
			t.Fatalf("expected snapshot to contain %s, got: %s", expected, snapshot)
		}
	}
}

func TestBuildPluginResourceRefDescriptorsDoNotPersistConcreteFilePaths(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-resource-summary")

	slotDir := filepath.Join(pluginDir, "frontend", "slots", "dashboard.workspace.after")
	if err := os.MkdirAll(slotDir, 0o755); err != nil {
		t.Fatalf("failed to create slot dir: %v", err)
	}
	writeTestFile(t, filepath.Join(slotDir, "workspace-card.vue"), "<template><div /></template>\n")

	descriptors := service.buildPluginResourceRefDescriptors(&pluginManifest{
		ID:           "plugin-resource-summary",
		Name:         "Resource Summary Plugin",
		Version:      "0.1.0",
		Type:         pluginTypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	})
	if len(descriptors) == 0 {
		t.Fatalf("expected resource descriptors to be generated")
	}

	for _, descriptor := range descriptors {
		if descriptor == nil {
			continue
		}
		if strings.Contains(descriptor.Key, "/") || strings.Contains(descriptor.OwnerKey, "/") {
			t.Fatalf("expected abstract resource identifiers without concrete file paths, got %#v", descriptor)
		}
		if strings.Contains(descriptor.Remark, ".vue") || strings.Contains(descriptor.Remark, ".sql") {
			t.Fatalf("expected remark to summarize resources without concrete file paths, got %#v", descriptor)
		}
	}
}

func TestBuildPluginResourceRefDescriptorsSummarizeRuntimeArtifact(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-resource-summary",
		"Runtime Resource Summary Plugin",
		"v0.5.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-resource-summary.sql", Content: "SELECT 1;"},
		},
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-resource-summary.sql", Content: "SELECT 2;"},
		},
	)

	manifest := &pluginManifest{
		ID:           "plugin-dynamic-resource-summary",
		Name:         "Runtime Resource Summary Plugin",
		Version:      "v0.5.0",
		Type:         pluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := service.validateRuntimePluginArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	descriptors := service.buildPluginResourceRefDescriptors(manifest)
	foundRuntimeArtifact := false
	for _, descriptor := range descriptors {
		if descriptor == nil {
			continue
		}
		if descriptor.Kind == pluginResourceKindRuntimeWasm {
			foundRuntimeArtifact = true
			if !strings.Contains(descriptor.Remark, "ABI v1") {
				t.Fatalf("expected dynamic artifact remark to mention ABI version, got %#v", descriptor)
			}
		}
	}
	if !foundRuntimeArtifact {
		t.Fatalf("expected runtime wasm descriptor to be generated")
	}
}

func TestResolvePluginSQLAssetsPrefersEmbeddedRuntimeSQL(t *testing.T) {
	service := New()
	pluginDir := createTestRuntimePluginDir(
		t,
		"plugin-dynamic-sql-assets",
		"Runtime SQL Assets Plugin",
		"v0.6.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-sql-assets.sql", Content: "SELECT 1;"},
			{Key: "002-plugin-dynamic-sql-assets.sql", Content: "SELECT 2;"},
		},
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-sql-assets.sql", Content: "SELECT 3;"},
		},
	)

	manifest := &pluginManifest{
		ID:           "plugin-dynamic-sql-assets",
		Name:         "Runtime SQL Assets Plugin",
		Version:      "v0.6.0",
		Type:         pluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := service.validateRuntimePluginArtifact(manifest, pluginDir); err != nil {
		t.Fatalf("expected dynamic artifact to be valid, got error: %v", err)
	}

	installAssets, err := service.resolvePluginSQLAssets(manifest, pluginMigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected install sql assets, got error: %v", err)
	}
	if len(installAssets) != 2 || installAssets[0].Key != "001-plugin-dynamic-sql-assets.sql" {
		t.Fatalf("unexpected install assets: %#v", installAssets)
	}

	uninstallAssets, err := service.resolvePluginSQLAssets(manifest, pluginMigrationDirectionUninstall)
	if err != nil {
		t.Fatalf("expected uninstall sql assets, got error: %v", err)
	}
	if len(uninstallAssets) != 1 || uninstallAssets[0].Content != "SELECT 3;" {
		t.Fatalf("unexpected uninstall assets: %#v", uninstallAssets)
	}
}

func TestResolvePluginSQLAssetsFallsBackToDirectoryConvention(t *testing.T) {
	service := New()
	pluginDir := createTestPluginDir(t, "plugin-directory-sql-assets")

	manifest := &pluginManifest{
		ID:           "plugin-directory-sql-assets",
		Name:         "Directory SQL Assets Plugin",
		Version:      "0.1.0",
		Type:         pluginTypeSource.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}

	installAssets, err := service.resolvePluginSQLAssets(manifest, pluginMigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected directory install sql assets, got error: %v", err)
	}
	if len(installAssets) != 1 || installAssets[0].Key != "001-plugin-directory-sql-assets.sql" {
		t.Fatalf("unexpected directory install assets: %#v", installAssets)
	}
}

func TestScanEmbeddedSourcePluginManifestsUsesPluginEmbeddedFiles(t *testing.T) {
	service := New()

	const pluginID = "plugin-embedded-manifest"
	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.UseEmbeddedFiles(fstest.MapFS{
		"plugin.yaml":                                 &fstest.MapFile{Data: []byte("id: plugin-embedded-manifest\nname: Embedded Manifest Plugin\nversion: 0.1.0\ntype: source\n")},
		"frontend/pages/main-entry.vue":              &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"frontend/slots/layout.header.after/tip.vue": &fstest.MapFile{Data: []byte("<template><div /></template>\n")},
		"manifest/sql/001-plugin-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 1;\n"),
		},
		"manifest/sql/uninstall/001-plugin-embedded-manifest.sql": &fstest.MapFile{
			Data: []byte("SELECT 2;\n"),
		},
	})
	pluginhost.RegisterSourcePlugin(sourcePlugin)

	manifests, err := service.scanEmbeddedSourcePluginManifests()
	if err != nil {
		t.Fatalf("expected embedded source manifests to load, got error: %v", err)
	}

	var target *pluginManifest
	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			target = manifest
			break
		}
	}
	if target == nil {
		t.Fatalf("expected embedded source plugin %s to be discovered", pluginID)
	}
	if target.ManifestPath != "embedded/source-plugins/plugin-embedded-manifest/plugin.yaml" {
		t.Fatalf("unexpected embedded manifest path: %s", target.ManifestPath)
	}
	if len(service.listPluginFrontendPagePaths(target)) != 1 {
		t.Fatalf("expected embedded frontend page paths to be discovered")
	}
	if len(service.listPluginFrontendSlotPaths(target)) != 1 {
		t.Fatalf("expected embedded frontend slot paths to be discovered")
	}
}

func TestResolvePluginSQLAssetsUsesEmbeddedSourcePluginFiles(t *testing.T) {
	service := New()

	manifest := &pluginManifest{
		ID:      "plugin-embedded-sql-assets",
		Name:    "Embedded SQL Assets Plugin",
		Version: "0.1.0",
		Type:    pluginTypeSource.String(),
		SourcePlugin: func() *pluginhost.SourcePlugin {
			sourcePlugin := pluginhost.NewSourcePlugin("plugin-embedded-sql-assets")
			sourcePlugin.UseEmbeddedFiles(fstest.MapFS{
				"plugin.yaml": &fstest.MapFile{Data: []byte("id: plugin-embedded-sql-assets\nname: Embedded SQL Assets Plugin\nversion: 0.1.0\ntype: source\n")},
				"manifest/sql/001-plugin-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 1;\n"),
				},
				"manifest/sql/uninstall/001-plugin-embedded-sql-assets.sql": &fstest.MapFile{
					Data: []byte("SELECT 2;\n"),
				},
			})
			return sourcePlugin
		}(),
	}

	installAssets, err := service.resolvePluginSQLAssets(manifest, pluginMigrationDirectionInstall)
	if err != nil {
		t.Fatalf("expected embedded install sql assets, got error: %v", err)
	}
	if len(installAssets) != 1 || installAssets[0].Content != "SELECT 1;" {
		t.Fatalf("unexpected embedded install assets: %#v", installAssets)
	}

	uninstallAssets, err := service.resolvePluginSQLAssets(manifest, pluginMigrationDirectionUninstall)
	if err != nil {
		t.Fatalf("expected embedded uninstall sql assets, got error: %v", err)
	}
	if len(uninstallAssets) != 1 || uninstallAssets[0].Content != "SELECT 2;" {
		t.Fatalf("unexpected embedded uninstall assets: %#v", uninstallAssets)
	}
}

func TestDerivePluginLifecycleState(t *testing.T) {
	testCases := []struct {
		name       string
		pluginType string
		installed  int
		enabled    int
		expected   string
	}{
		{
			name:       "source enabled",
			pluginType: pluginTypeSource.String(),
			installed:  pluginInstalledYes,
			enabled:    pluginStatusEnabled,
			expected:   pluginLifecycleStateSourceEnabled.String(),
		},
		{
			name:       "source disabled",
			pluginType: pluginTypeSource.String(),
			installed:  pluginInstalledYes,
			enabled:    pluginStatusDisabled,
			expected:   pluginLifecycleStateSourceDisabled.String(),
		},
		{
			name:       "runtime uninstalled",
			pluginType: pluginTypeDynamic.String(),
			installed:  pluginInstalledNo,
			enabled:    pluginStatusDisabled,
			expected:   pluginLifecycleStateRuntimeUninstalled.String(),
		},
		{
			name:       "runtime installed disabled",
			pluginType: pluginTypeDynamic.String(),
			installed:  pluginInstalledYes,
			enabled:    pluginStatusDisabled,
			expected:   pluginLifecycleStateRuntimeInstalled.String(),
		},
		{
			name:       "runtime enabled",
			pluginType: pluginTypeDynamic.String(),
			installed:  pluginInstalledYes,
			enabled:    pluginStatusEnabled,
			expected:   pluginLifecycleStateRuntimeEnabled.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := derivePluginLifecycleState(testCase.pluginType, testCase.installed, testCase.enabled)
			if actual != testCase.expected {
				t.Fatalf("expected lifecycle state %s, got %s", testCase.expected, actual)
			}
		})
	}
}

func TestDerivePluginNodeState(t *testing.T) {
	testCases := []struct {
		name      string
		installed int
		enabled   int
		expected  string
	}{
		{
			name:      "node uninstalled",
			installed: pluginInstalledNo,
			enabled:   pluginStatusDisabled,
			expected:  pluginNodeStateUninstalled.String(),
		},
		{
			name:      "node installed",
			installed: pluginInstalledYes,
			enabled:   pluginStatusDisabled,
			expected:  pluginNodeStateInstalled.String(),
		},
		{
			name:      "node enabled",
			installed: pluginInstalledYes,
			enabled:   pluginStatusEnabled,
			expected:  pluginNodeStateEnabled.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := derivePluginNodeState(testCase.installed, testCase.enabled)
			if actual != testCase.expected {
				t.Fatalf("expected node state %s, got %s", testCase.expected, actual)
			}
		})
	}
}

func createTestPluginDir(t *testing.T, pluginID string) string {
	t.Helper()

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID)
	if err = os.MkdirAll(filepath.Join(pluginDir, "backend"), 0o755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(pluginDir, "frontend", "pages"), 0o755); err != nil {
		t.Fatalf("failed to create frontend pages dir: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(pluginDir, "manifest", "sql", "uninstall"), 0o755); err != nil {
		t.Fatalf("failed to create sql dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(pluginDir)
	})

	writeTestFile(t, filepath.Join(pluginDir, "go.mod"), "module "+strings.ReplaceAll(pluginID, "-", "_")+"\n\ngo 1.25.0\n")
	writeTestFile(t, filepath.Join(pluginDir, "backend", "plugin.go"), "package backend\n")
	writeTestFile(t, filepath.Join(pluginDir, "frontend", "pages", "main-entry.vue"), "<template><div /></template>\n")
	writeTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"), "SELECT 1;\n")
	writeTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-"+pluginID+".sql"), "SELECT 1;\n")
	writeTestFile(t, filepath.Join(pluginDir, "plugin.yaml"), "id: "+pluginID+"\nname: test\nversion: 0.1.0\ntype: source\n")

	return pluginDir
}

func createTestRuntimePluginDir(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) string {
	return createTestRuntimePluginDirWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		defaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
}

func createTestRuntimeStorageArtifact(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) string {
	return createTestRuntimeStorageArtifactWithFilename(
		t,
		buildPluginDynamicArtifactFileName(pluginID),
		pluginID,
		pluginName,
		version,
		installSQLAssets,
		uninstallSQLAssets,
	)
}

func createTestRuntimeStorageArtifactWithFilename(
	t *testing.T,
	fileName string,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
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

	artifactPath := filepath.Join(storageDir, fileName)
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
			ABIVersion:         pluginDynamicSupportedABIVersion,
			FrontendAssetCount: len(defaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		defaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
	return artifactPath
}

func createTestRuntimePluginDirWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) string {
	t.Helper()

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID)
	if err = os.MkdirAll(filepath.Join(pluginDir, "runtime"), 0o755); err != nil {
		t.Fatalf("failed to create runtime dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(pluginDir)
	})

	writeTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\nname: "+pluginName+"\nversion: "+version+"\ntype: dynamic\n",
	)
	writeRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, buildPluginDynamicArtifactRelativePath(pluginID)),
		&pluginDynamicArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginDynamicSupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
	)

	return pluginDir
}

func defaultTestRuntimeFrontendAssets() []*pluginDynamicArtifactFrontendAsset {
	return []*pluginDynamicArtifactFrontendAsset{
		{
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
		{
			Path:          "assets/app.js",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend');")),
			ContentType:   "application/javascript",
		},
	}
}

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file %s: %v", filePath, err)
	}
}

func writeRuntimeWasmArtifact(
	t *testing.T,
	filePath string,
	manifest *pluginDynamicArtifactManifest,
	runtimeMetadata *pluginDynamicArtifactMetadata,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) {
	t.Helper()

	wasm := buildTestRuntimeWasmArtifactContent(
		t,
		manifest,
		runtimeMetadata,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
	)
	if err := os.WriteFile(filePath, wasm, 0o644); err != nil {
		t.Fatalf("failed to write runtime wasm artifact %s: %v", filePath, err)
	}
}

func buildTestRuntimeWasmArtifactContent(
	t *testing.T,
	manifest *pluginDynamicArtifactManifest,
	runtimeMetadata *pluginDynamicArtifactMetadata,
	frontendAssets []*pluginDynamicArtifactFrontendAsset,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) []byte {
	t.Helper()

	manifestContent, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal dynamic manifest: %v", err)
	}
	runtimeContent, err := json.Marshal(runtimeMetadata)
	if err != nil {
		t.Fatalf("failed to marshal runtime metadata: %v", err)
	}

	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendWasmCustomSection(wasm, pluginDynamicWasmSectionManifest, manifestContent)
	wasm = appendWasmCustomSection(wasm, pluginDynamicWasmSectionDynamic, runtimeContent)
	if len(frontendAssets) > 0 {
		frontendContent, err := json.Marshal(frontendAssets)
		if err != nil {
			t.Fatalf("failed to marshal frontend assets: %v", err)
		}
		wasm = appendWasmCustomSection(wasm, pluginDynamicWasmSectionFrontend, frontendContent)
	}
	if len(installSQLAssets) > 0 {
		installContent, err := json.Marshal(installSQLAssets)
		if err != nil {
			t.Fatalf("failed to marshal install sql assets: %v", err)
		}
		wasm = appendWasmCustomSection(wasm, pluginDynamicWasmSectionInstallSQL, installContent)
	}
	if len(uninstallSQLAssets) > 0 {
		uninstallContent, err := json.Marshal(uninstallSQLAssets)
		if err != nil {
			t.Fatalf("failed to marshal uninstall sql assets: %v", err)
		}
		wasm = appendWasmCustomSection(wasm, pluginDynamicWasmSectionUninstallSQL, uninstallContent)
	}
	return wasm
}
