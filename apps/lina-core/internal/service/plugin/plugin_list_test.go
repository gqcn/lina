// This file covers root-facade list methods defined in plugin_list.go.

package plugin

import (
	"context"
	"os"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-registry-missing"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Runtime Registry Missing Plugin",
		"v0.9.4",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to tolerate missing dynamic artifact, got error: %v", err)
	}

	var item *PluginItem
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			item = current
			break
		}
	}
	if item == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in plugin list")
	}
	if item.Installed != catalog.InstalledNo {
		t.Fatalf("expected missing dynamic plugin installed state to reconcile to %d, got %d", catalog.InstalledNo, item.Installed)
	}
	if item.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected missing dynamic plugin enabled state to reconcile to %d, got %d", catalog.StatusDisabled, item.Enabled)
	}

	runtimeStates, err := service.ListRuntimeStates(ctx)
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}
	var runtimeState *PluginDynamicStateItem
	for _, current := range runtimeStates.List {
		if current != nil && current.Id == pluginID {
			runtimeState = current
			break
		}
	}
	if runtimeState == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in public runtime states")
	}
	if runtimeState.Installed != catalog.InstalledNo || runtimeState.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected public runtime state to reconcile to uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after reconciliation")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected runtime registry row to reconcile to uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
}
