package plugin

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

func TestSingleNodeModeSkipsPluginNodeProjection(t *testing.T) {
	service := New()
	ctx := context.Background()

	var (
		pluginID   = "plugin-dynamic-single-node"
		pluginName = "Dynamic Single Node Plugin"
		version    = "v0.1.0"
	)

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	createTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		buildVersionedRuntimeFrontendAssets("single-node"),
		nil,
		nil,
	)

	if err := service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected single-node install to succeed, got error: %v", err)
	}
	if err := service.Enable(ctx, pluginID); err != nil {
		t.Fatalf("expected single-node enable to succeed, got error: %v", err)
	}

	nodeStateCount, err := dao.SysPluginNodeState.Ctx(ctx).
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected plugin node-state count query to succeed, got error: %v", err)
	}
	if nodeStateCount != 0 {
		t.Fatalf("expected single-node mode to skip node-state projection rows, got %d", nodeStateCount)
	}

	snapshot, err := service.buildPluginGovernanceSnapshot(
		ctx,
		pluginID,
		version,
		pluginTypeDynamic.String(),
		pluginInstalledYes,
		pluginStatusEnabled,
	)
	if err != nil {
		t.Fatalf("expected governance snapshot build to succeed, got error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected governance snapshot to exist")
	}
	if snapshot.NodeState != pluginNodeStateEnabled.String() {
		t.Fatalf("expected governance snapshot to derive enabled node state, got %s", snapshot.NodeState)
	}
}
