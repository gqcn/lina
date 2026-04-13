package plugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/pluginbridge"
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

func TestBundledDynamicPluginEnableMakesDynamicRouteExecutable(t *testing.T) {
	service := New()
	ctx := context.Background()

	const pluginID = "plugin-demo-dynamic"

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("expected repo root resolution to succeed, got error: %v", err)
	}
	cmd := exec.Command("make", "wasm", "p=plugin-demo-dynamic", "out=../../temp/output")
	cmd.Dir = filepath.Join(repoRoot, "apps", "lina-plugins")
	if output, buildErr := cmd.CombinedOutput(); buildErr != nil {
		t.Fatalf("expected bundled dynamic plugin artifact build to succeed, got error: %v output=%s", buildErr, string(output))
	}

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected bundled dynamic plugin install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected bundled dynamic plugin enable to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected plugin registry row after enable")
	}
	if registry.Installed != pluginInstalledYes || registry.Status != pluginStatusEnabled {
		t.Fatalf("expected bundled dynamic plugin to be installed+enabled, got %#v", registry)
	}
	if registry.ReleaseId <= 0 {
		t.Fatalf("expected bundled dynamic plugin to keep active release id, got %d", registry.ReleaseId)
	}

	manifest, err := service.getActivePluginManifest(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected active plugin manifest to load, got error: %v", err)
	}
	if manifest == nil || len(manifest.Routes) == 0 || manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		t.Fatalf("expected bundled dynamic plugin active manifest to expose executable routes, got %#v", manifest)
	}
	if manifest.Routes[0].Path != "/backend-summary" || manifest.Routes[0].Method != http.MethodGet {
		t.Fatalf("expected bundled dynamic plugin active route to expose GET /backend-summary, got %#v", manifest.Routes[0])
	}

	response, err := service.executeDynamicRoute(ctx, manifest, &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: pluginID,
		Route: &pluginbridge.RouteMatchSnapshotV1{
			InternalPath: "/backend-summary",
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/backend-summary",
			Access:       pluginbridge.AccessLogin,
			Permission:   "plugin-demo-dynamic:backend:view",
		},
		Identity: &pluginbridge.IdentitySnapshotV1{
			UserID:       1,
			Username:     "admin",
			IsSuperAdmin: true,
		},
		Request: &pluginbridge.HTTPRequestSnapshotV1{
			Method: http.MethodGet,
		},
	})
	if err != nil {
		t.Fatalf("expected bundled dynamic plugin route execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected bundled dynamic plugin route response 200, got %#v", response)
	}

	payload := map[string]any{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected bundled dynamic plugin response to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != pluginID {
		t.Fatalf("expected bundled dynamic plugin payload to preserve pluginId, got %#v", payload)
	}
}

func TestInstallSameVersionDynamicPluginRefreshesArchivedReleaseArtifact(t *testing.T) {
	service := New()
	ctx := context.Background()

	pluginID := "plugin-dynamic-same-version-refresh"
	pluginName := "Dynamic Same Version Refresh Plugin"
	version := "v0.1.0"

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	initialRoutes := []*pluginbridge.RouteContract{
		{
			Path:       "/review-summary",
			Method:     http.MethodGet,
			Access:     pluginbridge.AccessLogin,
			Permission: pluginID + ":review:view",
			OperLog:    "other",
		},
	}
	initialBridge := &pluginbridge.BridgeSpec{
		ABIVersion:     pluginbridge.ABIVersionV1,
		RuntimeKind:    pluginbridge.RuntimeKindWasm,
		RouteExecution: true,
		RequestCodec:   pluginbridge.CodecProtobuf,
		ResponseCodec:  pluginbridge.CodecProtobuf,
		AllocExport:    pluginbridge.DefaultGuestAllocExport,
		ExecuteExport:  pluginbridge.DefaultGuestExecuteExport,
	}
	createTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("version-one"),
		nil,
		nil,
		initialRoutes,
		initialBridge,
	)

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected initial enable to succeed, got error: %v", err)
	}

	registryBeforeRefresh, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup before refresh to succeed, got error: %v", err)
	}
	if registryBeforeRefresh == nil {
		t.Fatal("expected registry row before same-version refresh")
	}

	refreshedRoutes := []*pluginbridge.RouteContract{
		{
			Path:       "/review-summary",
			Method:     http.MethodGet,
			Access:     pluginbridge.AccessLogin,
			Permission: pluginID + ":review:inspect",
			OperLog:    "other",
		},
	}
	createTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("version-two"),
		nil,
		nil,
		refreshedRoutes,
		initialBridge,
	)

	if err = service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected same-version refresh install to succeed, got error: %v", err)
	}

	registryAfterRefresh, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after refresh to succeed, got error: %v", err)
	}
	if registryAfterRefresh == nil {
		t.Fatal("expected registry row after same-version refresh")
	}
	if registryAfterRefresh.ReleaseId != registryBeforeRefresh.ReleaseId {
		t.Fatalf("expected same-version refresh to reuse active release id, before=%d after=%d", registryBeforeRefresh.ReleaseId, registryAfterRefresh.ReleaseId)
	}
	if registryAfterRefresh.Generation <= registryBeforeRefresh.Generation {
		t.Fatalf("expected same-version refresh to advance generation, before=%d after=%d", registryBeforeRefresh.Generation, registryAfterRefresh.Generation)
	}

	activeManifest, err := service.getActivePluginManifest(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected active manifest after refresh to load, got error: %v", err)
	}
	if activeManifest == nil || activeManifest.RuntimeArtifact == nil {
		t.Fatalf("expected active manifest runtime artifact after refresh, got %#v", activeManifest)
	}
	if len(activeManifest.Routes) != 1 || activeManifest.Routes[0].Permission != pluginID+":review:inspect" {
		t.Fatalf("expected active manifest routes to refresh with new permission, got %#v", activeManifest.Routes)
	}

	asset, err := service.ResolveRuntimeFrontendAsset(ctx, pluginID, version, "index.html")
	if err != nil {
		t.Fatalf("expected refreshed frontend asset to resolve, got error: %v", err)
	}
	if !strings.Contains(string(asset.Content), "version-two") {
		t.Fatalf("expected refreshed frontend asset to contain version-two marker, got %s", string(asset.Content))
	}
}
