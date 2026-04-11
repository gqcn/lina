package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/pluginhost"
)

func TestBuildRuntimeWasmArtifactEmbedsBackendContracts(t *testing.T) {
	service := New()
	pluginDir := t.TempDir()

	mustWriteRuntimeSourceFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-contract\nname: Dynamic Contract\nversion: v0.2.0\ntype: dynamic\n",
	)
	mustWriteRuntimeSourceFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		strings.Join([]string{
			"event: auth.login.succeeded",
			"action: sleep",
			"timeoutMs: 50",
			"sleepMs: 10",
		}, "\n"),
	)
	mustWriteRuntimeSourceFile(
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

	buildOut, err := service.BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact build to succeed, got error: %v", err)
	}

	artifact, err := service.parseRuntimeWasmArtifactContent(buildOut.ArtifactPath, buildOut.Content)
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
	service := New()
	pluginDir := t.TempDir()

	mustWriteRuntimeSourceFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-active-contract\nname: Active Contract\nversion: v0.2.0\ntype: dynamic\n",
	)
	mustWriteRuntimeSourceFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		strings.Join([]string{
			"event: auth.login.succeeded",
			"action: sleep",
			"timeoutMs: 50",
			"sleepMs: 10",
		}, "\n"),
	)
	mustWriteRuntimeSourceFile(
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

	buildOut, err := service.BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact build to succeed, got error: %v", err)
	}
	if err = os.MkdirAll(filepath.Dir(buildOut.ArtifactPath), 0o755); err != nil {
		t.Fatalf("expected runtime artifact directory to be created, got error: %v", err)
	}
	if err = os.WriteFile(buildOut.ArtifactPath, buildOut.Content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact to be written, got error: %v", err)
	}

	manifest, err := service.loadRuntimePluginManifestFromArtifact(buildOut.ArtifactPath)
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

func TestRunPluginDeclaredHookHonorsTimeoutAndErrorActions(t *testing.T) {
	service := New()

	sleepHook := &pluginHookSpec{
		Event:     pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:    pluginhost.HookActionSleep,
		TimeoutMs: 10,
		SleepMs:   80,
	}
	timeoutCtx, cancel := service.buildPluginHookTimeoutContext(context.Background(), sleepHook)
	defer cancel()

	err := service.runPluginDeclaredHook(timeoutCtx, "plugin-dynamic-timeout", sleepHook, nil)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error for sleep hook, got: %v", err)
	}

	errorHook := &pluginHookSpec{
		Event:        pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:       pluginhost.HookActionError,
		ErrorMessage: "runtime hook failed on purpose",
	}
	err = service.runPluginDeclaredHook(context.Background(), "plugin-dynamic-error", errorHook, nil)
	if err == nil || !strings.Contains(err.Error(), "runtime hook failed on purpose") {
		t.Fatalf("expected declared error hook message, got: %v", err)
	}
}

func mustWriteRuntimeSourceFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create source directory %s: %v", filePath, err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write source file %s: %v", filePath, err)
	}
}
