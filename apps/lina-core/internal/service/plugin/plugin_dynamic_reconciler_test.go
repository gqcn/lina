package plugin

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
)

func TestDynamicPluginUpgradeKeepsPreviousReleaseFrontendAssets(t *testing.T) {
	service := New()
	ctx := context.Background()

	pluginID := "plugin-dynamic-upgrade"
	pluginName := "Dynamic Upgrade Plugin"
	versionOne := "v0.1.0"
	versionTwo := "v0.2.0"

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("version-one"),
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeUpgrade, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registryBeforeUpgrade == nil {
		t.Fatal("expected registry row to exist after initial enable")
	}

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionTwo,
		buildVersionedRuntimeFrontendAssets("version-two"),
		nil,
		nil,
	)

	if err = service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected upgrade install to succeed, got error: %v", err)
	}

	registryAfterUpgrade, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected upgraded registry lookup to succeed, got error: %v", err)
	}
	if registryAfterUpgrade == nil {
		t.Fatal("expected upgraded registry row to exist")
	}
	if registryAfterUpgrade.Version != versionTwo {
		t.Fatalf("expected active version %s after upgrade, got %s", versionTwo, registryAfterUpgrade.Version)
	}
	if registryAfterUpgrade.Generation <= registryBeforeUpgrade.Generation {
		t.Fatalf("expected generation to advance after upgrade, before=%d after=%d", registryBeforeUpgrade.Generation, registryAfterUpgrade.Generation)
	}
	if registryAfterUpgrade.ReleaseId == registryBeforeUpgrade.ReleaseId {
		t.Fatalf("expected active release id to change after upgrade, got %d", registryAfterUpgrade.ReleaseId)
	}

	oldAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionOne, "index.html")
	if err != nil {
		t.Fatalf("expected previous release asset to stay resolvable, got error: %v", err)
	}
	if !strings.Contains(string(oldAsset.Content), "version-one") {
		t.Fatalf("expected previous release asset content to contain version-one marker, got %s", string(oldAsset.Content))
	}

	newAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionTwo, "index.html")
	if err != nil {
		t.Fatalf("expected new release asset to be resolvable, got error: %v", err)
	}
	if !strings.Contains(string(newAsset.Content), "version-two") {
		t.Fatalf("expected new release asset content to contain version-two marker, got %s", string(newAsset.Content))
	}

	releaseOne, err := service.getPluginRelease(ctx, pluginID, versionOne)
	if err != nil {
		t.Fatalf("expected previous release lookup to succeed, got error: %v", err)
	}
	releaseTwo, err := service.getPluginRelease(ctx, pluginID, versionTwo)
	if err != nil {
		t.Fatalf("expected new release lookup to succeed, got error: %v", err)
	}
	if releaseOne == nil || releaseOne.Status != pluginReleaseStatusInstalled.String() {
		t.Fatalf("expected previous release to remain installed for drain/rollback, got %#v", releaseOne)
	}
	if releaseTwo == nil || releaseTwo.Status != pluginReleaseStatusActive.String() {
		t.Fatalf("expected new release to become active, got %#v", releaseTwo)
	}
}

func TestDynamicPluginUpgradeFailureRollsBackStableRelease(t *testing.T) {
	service := New()
	ctx := context.Background()

	pluginID := "plugin-dynamic-upgrade-failed"
	pluginName := "Dynamic Upgrade Failure Plugin"
	versionOne := "v0.1.0"
	versionTwo := "v0.2.0"

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("stable-version"),
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registryBeforeFailure == nil {
		t.Fatal("expected registry row before failed upgrade")
	}

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionTwo,
		buildVersionedRuntimeFrontendAssets("broken-version"),
		[]*pluginDynamicArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-upgrade-failed.sql",
				Content: "THIS IS NOT VALID SQL;",
			},
		},
		nil,
	)

	if err = service.Install(ctx, pluginID); err == nil {
		t.Fatal("expected failed upgrade install to return an error")
	}

	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil {
		t.Fatal("expected registry row after failed upgrade")
	}
	if registryAfterFailure.Version != versionOne {
		t.Fatalf("expected active version to stay at %s after rollback, got %s", versionOne, registryAfterFailure.Version)
	}
	if registryAfterFailure.ReleaseId != registryBeforeFailure.ReleaseId {
		t.Fatalf("expected active release id to stay unchanged after rollback, before=%d after=%d", registryBeforeFailure.ReleaseId, registryAfterFailure.ReleaseId)
	}
	if registryAfterFailure.Generation != registryBeforeFailure.Generation {
		t.Fatalf("expected generation to stay unchanged after rollback, before=%d after=%d", registryBeforeFailure.Generation, registryAfterFailure.Generation)
	}
	if registryAfterFailure.DesiredState != pluginHostStateEnabled.String() || registryAfterFailure.CurrentState != pluginHostStateEnabled.String() {
		t.Fatalf("expected registry to restore enabled stable state after rollback, got desired=%s current=%s", registryAfterFailure.DesiredState, registryAfterFailure.CurrentState)
	}

	stableAsset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionOne, "index.html")
	if err != nil {
		t.Fatalf("expected stable release asset to remain resolvable after rollback, got error: %v", err)
	}
	if !strings.Contains(string(stableAsset.Content), "stable-version") {
		t.Fatalf("expected stable release asset content to be preserved, got %s", string(stableAsset.Content))
	}

	failedRelease, err := service.getPluginRelease(ctx, pluginID, versionTwo)
	if err != nil {
		t.Fatalf("expected failed release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil || failedRelease.Status != pluginReleaseStatusFailed.String() {
		t.Fatalf("expected failed release status to be marked failed, got %#v", failedRelease)
	}
	if _, err = service.ResolveRuntimeFrontendAsset(ctx, pluginID, versionTwo, "index.html"); err == nil {
		t.Fatal("expected failed release asset to stay hidden from runtime frontend resolution")
	}
}

func TestDynamicPluginFollowerDefersUntilPrimaryReconciles(t *testing.T) {
	topology := &testTopology{
		enabled: true,
		primary: false,
		nodeID:  "follower-node",
	}
	service := New(topology)
	ctx := context.Background()

	pluginID := "plugin-dynamic-follower"
	pluginName := "Dynamic Follower Plugin"
	versionOne := "v0.1.0"

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		versionOne,
		buildVersionedRuntimeFrontendAssets("follower-version"),
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected follower-side install request to persist desired state, got error: %v", err)
	}

	registryBeforePrimary, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected follower registry lookup to succeed, got error: %v", err)
	}
	if registryBeforePrimary == nil {
		t.Fatal("expected registry row to exist on follower")
	}
	if registryBeforePrimary.Installed != pluginInstalledNo {
		t.Fatalf("expected follower request to keep current install state unchanged, got installed=%d", registryBeforePrimary.Installed)
	}
	if registryBeforePrimary.DesiredState != pluginHostStateInstalled.String() {
		t.Fatalf("expected follower request to persist desired installed state, got %s", registryBeforePrimary.DesiredState)
	}
	if registryBeforePrimary.CurrentState != pluginHostStateUninstalled.String() {
		t.Fatalf("expected follower current state to remain uninstalled before primary reconciliation, got %s", registryBeforePrimary.CurrentState)
	}

	topology.primary = true
	if err = service.ReconcileRuntimePlugins(ctx); err != nil {
		t.Fatalf("expected primary reconciliation to succeed, got error: %v", err)
	}

	registryAfterPrimary, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected primary registry lookup to succeed, got error: %v", err)
	}
	if registryAfterPrimary == nil {
		t.Fatal("expected registry row after primary reconciliation")
	}
	if registryAfterPrimary.Installed != pluginInstalledYes {
		t.Fatalf("expected primary reconciliation to install plugin, got installed=%d", registryAfterPrimary.Installed)
	}
	if registryAfterPrimary.CurrentState != pluginHostStateInstalled.String() {
		t.Fatalf("expected current state to converge to installed on primary, got %s", registryAfterPrimary.CurrentState)
	}
	if registryAfterPrimary.ReleaseId <= 0 {
		t.Fatalf("expected primary reconciliation to persist active release id, got %d", registryAfterPrimary.ReleaseId)
	}
}

func buildVersionedRuntimeFrontendAssets(marker string) []*pluginDynamicArtifactFrontendAsset {
	return []*pluginDynamicArtifactFrontendAsset{
		{
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>" + marker + "</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
	}
}
