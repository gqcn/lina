// This file covers root-facade lifecycle methods defined in plugin_lifecycle.go.

package plugin

import (
	"context"
	"os"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/testutil"
)

func TestUpdateStatusEnablesBackendOnlyDynamicPluginWithoutFrontendAssets(t *testing.T) {
	var (
		service  = New()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-backend-only"
	)

	frontend.ResetBundleCache()
	t.Cleanup(frontend.ResetBundleCache)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Backend Only Dynamic Plugin",
		"v0.4.1",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		_ = os.Remove(artifactPath)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected backend-only artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected backend-only artifact sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected backend-only plugin install state to be set, got error: %v", err)
	}

	if err = service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected backend-only dynamic plugin enable to succeed, got error: %v", err)
	}
	if !service.IsEnabled(ctx, pluginID) {
		t.Fatalf("expected backend-only dynamic plugin to be enabled after status update")
	}
}
