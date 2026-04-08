// This file contains unit tests for manifest validation, convention-based
// resource discovery, and review-oriented plugin metadata helpers.

package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			pluginType: pluginTypeRuntime.String(),
			installed:  pluginInstalledNo,
			enabled:    pluginStatusDisabled,
			expected:   pluginLifecycleStateRuntimeUninstalled.String(),
		},
		{
			name:       "runtime installed disabled",
			pluginType: pluginTypeRuntime.String(),
			installed:  pluginInstalledYes,
			enabled:    pluginStatusDisabled,
			expected:   pluginLifecycleStateRuntimeInstalled.String(),
		},
		{
			name:       "runtime enabled",
			pluginType: pluginTypeRuntime.String(),
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

func writeTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file %s: %v", filePath, err)
	}
}
