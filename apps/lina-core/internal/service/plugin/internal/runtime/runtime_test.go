// This file covers backend contract sections embedded into runtime wasm artifacts.

package runtime_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginhost"
)

func TestBuildRuntimeWasmArtifactEmbedsBackendContracts(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := t.TempDir()

	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-contract\nname: Dynamic Contract\nversion: v0.2.0\ntype: dynamic\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		strings.Join([]string{
			"event: auth.login.succeeded",
			"action: sleep",
			"timeoutMs: 50",
			"sleepMs: 10",
		}, "\n"),
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "backend", "resources", "001-records.yaml"),
		strings.Join([]string{
			"key: records",
			"type: table-list",
			"table: plugin_runtime_records",
			"fields:",
			"  - name: id",
			"    column: id",
			"orderBy:",
			"  column: id",
			"  direction: asc",
			"dataScope:",
			"  userColumn: owner_user_id",
		}, "\n"),
	)

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, pluginDir)

	artifact, err := services.Runtime.ParseRuntimeWasmArtifactContent(buildOut.ArtifactPath, buildOut.Content)
	if err != nil {
		t.Fatalf("expected dynamic artifact parse to succeed, got error: %v", err)
	}
	if len(artifact.HookSpecs) != 1 {
		t.Fatalf("expected 1 embedded hook spec, got %d", len(artifact.HookSpecs))
	}
	if artifact.HookSpecs[0].Action != pluginhost.HookActionSleep {
		t.Fatalf("expected embedded hook action sleep, got %s", artifact.HookSpecs[0].Action)
	}
	if len(artifact.ResourceSpecs) != 1 {
		t.Fatalf("expected 1 embedded resource spec, got %d", len(artifact.ResourceSpecs))
	}
	if artifact.ResourceSpecs[0].DataScope == nil || artifact.ResourceSpecs[0].DataScope.UserColumn != "owner_user_id" {
		t.Fatalf("expected embedded resource data scope userColumn owner_user_id, got %#v", artifact.ResourceSpecs[0].DataScope)
	}
}

func TestLoadRuntimePluginManifestFromArtifactHydratesBackendContracts(t *testing.T) {
	services := testutil.NewServices()
	pluginDir := t.TempDir()

	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-active-contract\nname: Active Contract\nversion: v0.2.0\ntype: dynamic\n",
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		strings.Join([]string{
			"event: auth.login.succeeded",
			"action: sleep",
			"timeoutMs: 50",
			"sleepMs: 10",
		}, "\n"),
	)
	testutil.WriteTestFile(
		t,
		filepath.Join(pluginDir, "backend", "resources", "001-records.yaml"),
		strings.Join([]string{
			"key: records",
			"type: table-list",
			"table: plugin_runtime_records",
			"fields:",
			"  - name: id",
			"    column: id",
			"orderBy:",
			"  column: id",
			"  direction: asc",
		}, "\n"),
	)

	buildOut := testutil.BuildRuntimeArtifactWithHackTool(t, pluginDir)
	if err := os.MkdirAll(filepath.Dir(buildOut.ArtifactPath), 0o755); err != nil {
		t.Fatalf("expected runtime artifact directory to be created, got error: %v", err)
	}
	if err := os.WriteFile(buildOut.ArtifactPath, buildOut.Content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact to be written, got error: %v", err)
	}

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(buildOut.ArtifactPath)
	if err != nil {
		t.Fatalf("expected runtime manifest load to succeed, got error: %v", err)
	}
	if len(manifest.Hooks) != 1 {
		t.Fatalf("expected runtime manifest to expose 1 hook, got %d", len(manifest.Hooks))
	}
	if len(manifest.BackendResources) != 1 {
		t.Fatalf("expected runtime manifest to expose 1 backend resource, got %d", len(manifest.BackendResources))
	}
	if _, ok := manifest.BackendResources["records"]; !ok {
		t.Fatalf("expected runtime manifest to expose resource key records, got %#v", manifest.BackendResources)
	}
}
