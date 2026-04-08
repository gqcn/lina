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
		Type:        pluginTypeSource,
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
		Type:    pluginTypeSource,
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
