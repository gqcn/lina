package plugin

import (
	"context"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

func TestSyncPluginResourceReferencesRevivesSoftDeletedRows(t *testing.T) {
	service := New()
	ctx := context.Background()

	pluginID := "plugin-dynamic-ref-revive"
	pluginDir := createTestRuntimePluginDir(
		t,
		pluginID,
		"Runtime Ref Revive Plugin",
		"v0.9.0",
		[]*pluginDynamicArtifactSQLAsset{
			{Key: "001-plugin-dynamic-ref-revive.sql", Content: "SELECT 1;"},
		},
		nil,
	)

	manifest := &pluginManifest{
		ID:           pluginID,
		Name:         "Runtime Ref Revive Plugin",
		Version:      "v0.9.0",
		Type:         pluginTypeDynamic.String(),
		ManifestPath: filepath.Join(pluginDir, "plugin.yaml"),
		RootDir:      pluginDir,
	}
	if err := service.validatePluginManifest(manifest, manifest.ManifestPath); err != nil {
		t.Fatalf("expected dynamic manifest to be valid, got error: %v", err)
	}

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected plugin manifest sync to succeed, got error: %v", err)
	}

	release, err := service.getPluginRelease(ctx, pluginID, manifest.Version)
	if err != nil {
		t.Fatalf("expected plugin release to exist, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected plugin release to be created")
	}

	if _, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{
			PluginId:  pluginID,
			ReleaseId: release.Id,
		}).
		Delete(); err != nil {
		t.Fatalf("expected resource refs to be soft-deleted, got error: %v", err)
	}

	if err = service.syncPluginResourceReferences(ctx, manifest); err != nil {
		t.Fatalf("expected sync to revive soft-deleted rows without duplicate-key errors, got error: %v", err)
	}

	activeRefs, err := service.listPluginResourceRefs(ctx, pluginID, release.Id)
	if err != nil {
		t.Fatalf("expected resource refs to be queryable, got error: %v", err)
	}
	if len(activeRefs) == 0 {
		t.Fatalf("expected revived resource refs to exist")
	}

	for _, item := range activeRefs {
		if item == nil {
			continue
		}
		if item.DeletedAt != nil {
			t.Fatalf("expected revived resource ref %s/%s to be active, got deleted_at=%v", item.ResourceType, item.ResourceKey, item.DeletedAt)
		}
	}
}

func cleanupPluginGovernanceRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	_, _ = dao.SysPluginNodeState.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete()
	_, _ = dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Delete()
	_, _ = dao.SysPluginMigration.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginMigration{PluginId: pluginID}).
		Delete()
	_, _ = dao.SysPluginRelease.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginRelease{PluginId: pluginID}).
		Delete()
	_, _ = dao.SysPlugin.Ctx(ctx).
		Unscoped().
		Where(do.SysPlugin{PluginId: pluginID}).
		Delete()
}
