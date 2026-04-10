package plugin

import (
	"context"
	"os"
	"testing"

	"lina-core/internal/model/entity"
)

func TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-registry-missing"
	)

	artifactPath := createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Registry Missing Plugin",
		"v0.9.4",
		nil,
		nil,
	)

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, pluginStatusEnabled); err != nil {
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
	if item.Installed != pluginInstalledNo {
		t.Fatalf("expected missing dynamic plugin installed state to reconcile to %d, got %d", pluginInstalledNo, item.Installed)
	}
	if item.Enabled != pluginStatusDisabled {
		t.Fatalf("expected missing dynamic plugin enabled state to reconcile to %d, got %d", pluginStatusDisabled, item.Enabled)
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
	if runtimeState.Installed != pluginInstalledNo || runtimeState.Enabled != pluginStatusDisabled {
		t.Fatalf("expected public runtime state to reconcile to uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after reconciliation")
	}
	if registry.Installed != pluginInstalledNo || registry.Status != pluginStatusDisabled {
		t.Fatalf("expected runtime registry row to reconcile to uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
}

func TestFilterMenusHidesRuntimeMenusWhenArtifactIsMissing(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-menu-hidden"
	)

	artifactPath := createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Menu Hidden Plugin",
		"v0.9.5",
		nil,
		nil,
	)

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, pluginStatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	filtered := service.FilterMenus(ctx, []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "plugin:" + pluginID + ":entry",
			Name:    "runtime menu",
			Type:    "M",
			Status:  1,
			Visible: 1,
		},
	})
	if len(filtered) != 0 {
		t.Fatalf("expected dynamic plugin menu to be hidden after artifact removal, got %d entries", len(filtered))
	}
}

func TestUploadDynamicPackageAllowsRecoveryWhenArtifactIsMissing(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-upload-recover"
	)

	artifactPath := createTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Runtime Upload Recover Plugin",
		"v0.9.6",
		nil,
		nil,
	)

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read dynamic artifact content: %v", err)
	}

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, pluginInstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, pluginStatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.storeUploadedRuntimePackage(
		ctx,
		buildPluginDynamicArtifactFileName(pluginID),
		content,
		false,
	)
	if err != nil {
		t.Fatalf("expected runtime upload recovery to succeed, got error: %v", err)
	}
	if out.Installed != pluginInstalledNo {
		t.Fatalf("expected recovery upload to keep plugin uninstalled, got %d", out.Installed)
	}
	if out.Enabled != pluginStatusDisabled {
		t.Fatalf("expected recovery upload to keep plugin disabled, got %d", out.Enabled)
	}

	exists, _, err := service.hasRuntimeArtifactStorageFile(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected uploaded dynamic artifact lookup to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatalf("expected recovery upload to restore dynamic artifact into storage")
	}
}
