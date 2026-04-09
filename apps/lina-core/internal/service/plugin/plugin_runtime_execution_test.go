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
		"id: plugin-runtime-contract\nname: Runtime Contract\nversion: v0.2.0\ntype: runtime\n",
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
		t.Fatalf("expected runtime artifact build to succeed, got error: %v", err)
	}

	artifact, err := service.parseRuntimeWasmArtifactContent(buildOut.ArtifactPath, buildOut.Content)
	if err != nil {
		t.Fatalf("expected runtime artifact parse to succeed, got error: %v", err)
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

	err := service.runPluginDeclaredHook(timeoutCtx, "plugin-runtime-timeout", sleepHook, nil)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error for sleep hook, got: %v", err)
	}

	errorHook := &pluginHookSpec{
		Event:        pluginhost.ExtensionPointAuthLoginSucceeded,
		Action:       pluginhost.HookActionError,
		ErrorMessage: "runtime hook failed on purpose",
	}
	err = service.runPluginDeclaredHook(context.Background(), "plugin-runtime-error", errorHook, nil)
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
