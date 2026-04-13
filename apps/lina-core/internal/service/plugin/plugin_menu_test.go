package plugin

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/pluginbridge"
)

func TestSyncSourcePluginMenusFromManifest(t *testing.T) {
	service := New()
	ctx := context.Background()

	const (
		pluginID = "plugin-source-menu-sync"
		menuKey  = "plugin:plugin-source-menu-sync:sidebar-entry"
	)

	pluginDir := createTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: Source Menu Sync Plugin\n"+
			"version: v0.1.0\n"+
			"type: source\n"+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: Source Menu Sync Plugin\n"+
			"    path: plugin-source-menu-sync\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: plugin-source-menu-sync:view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)

	manifest := &pluginManifest{
		ID:      pluginID,
		Name:    "Source Menu Sync Plugin",
		Version: "v0.1.0",
		Type:    pluginTypeSource.String(),
		Menus: []*pluginMenuSpec{
			{
				Key:       menuKey,
				Name:      "Source Menu Sync Plugin",
				Path:      "plugin-source-menu-sync",
				Component: "system/plugin/dynamic-page",
				Perms:     "plugin-source-menu-sync:view",
				Icon:      "ant-design:appstore-outlined",
				Type:      pluginMenuTypePage.String(),
				Sort:      -1,
			},
		},
		ManifestPath: manifestPath,
		RootDir:      pluginDir,
	}
	if err := service.validatePluginManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected source plugin manifest with menus to be valid, got error: %v", err)
	}

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	cleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginMenuRowsHard(t, ctx, pluginID)
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected source plugin menu sync to succeed, got error: %v", err)
	}

	menu, err := queryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected source plugin menu %s to be created", menuKey)
	}
	if menu.Path != "plugin-source-menu-sync" {
		t.Fatalf("expected source plugin menu path to be synced, got %s", menu.Path)
	}

	roleMenuCount, err := dao.SysRoleMenu.Ctx(ctx).
		Where(do.SysRoleMenu{
			RoleId: pluginDefaultAdminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 1 {
		t.Fatalf("expected source plugin menu to be granted to admin role, got count=%d", roleMenuCount)
	}

	writeTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\nname: Source Menu Sync Plugin\nversion: v0.1.0\ntype: source\n",
	)
	manifest.Menus = nil
	if err := service.validatePluginManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected source plugin manifest without menus to stay valid, got error: %v", err)
	}
	if _, err := service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected source plugin stale menu cleanup to succeed, got error: %v", err)
	}

	menu, err = queryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu cleanup query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu %s to be deleted after manifest removed it", menuKey)
	}
}

func TestDynamicPluginInstallAndUninstallManageMenusFromManifest(t *testing.T) {
	service := New()
	ctx := context.Background()

	const (
		pluginID = "plugin-dynamic-menu-metadata"
		menuKey  = "plugin:plugin-dynamic-menu-metadata:main-entry"
	)

	artifactPath := createTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Menu Metadata Plugin",
		"v0.3.0",
		[]*pluginMenuSpec{
			{
				Key:       menuKey,
				Name:      "Runtime Menu Metadata Plugin",
				Path:      "/plugin-assets/plugin-dynamic-menu-metadata/v0.3.0/index.html",
				Perms:     "plugin-dynamic-menu-metadata:view",
				Icon:      "ant-design:deployment-unit-outlined",
				Type:      pluginMenuTypePage.String(),
				Sort:      -1,
				Query:     map[string]interface{}{"pluginAccessMode": "embedded-mount"},
				Component: "system/plugin/dynamic-page",
			},
		},
		nil,
		nil,
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact with manifest menus to load, got error: %v", err)
	}

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	cleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginMenuRowsHard(t, ctx, pluginID)
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected runtime plugin manifest sync to succeed, got error: %v", err)
	}
	if err = service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin install to succeed, got error: %v", err)
	}

	menu, err := queryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected runtime plugin menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected runtime plugin menu %s to be created on install", menuKey)
	}

	roleMenuCount, err := dao.SysRoleMenu.Ctx(ctx).
		Where(do.SysRoleMenu{
			RoleId: pluginDefaultAdminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected runtime admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 1 {
		t.Fatalf("expected runtime plugin menu to be granted to admin role, got count=%d", roleMenuCount)
	}

	if err = service.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin uninstall to succeed, got error: %v", err)
	}

	menu, err = queryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected runtime plugin menu cleanup query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected runtime plugin menu %s to be deleted on uninstall", menuKey)
	}
}

func TestDynamicPluginRoutePermissionsMaterializeHiddenMenus(t *testing.T) {
	service := New()
	ctx := context.Background()

	const pluginID = "plugin-dynamic-route-permission"

	artifactPath := createTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Route Permission Plugin",
		"v0.3.0",
		nil,
		nil,
		nil,
	)

	writeRuntimeWasmArtifact(
		t,
		artifactPath,
		&pluginDynamicArtifactManifest{
			ID:      pluginID,
			Name:    "Runtime Route Permission Plugin",
			Version: "v0.3.0",
			Type:    pluginTypeDynamic.String(),
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(defaultTestRuntimeFrontendAssets()),
			RouteCount:         1,
		},
		defaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		[]*pluginbridge.RouteContract{
			{
				Path:       "/review-summary",
				Method:     http.MethodGet,
				Access:     pluginbridge.AccessLogin,
				Permission: "plugin-dynamic-route-permission:review:view",
			},
		},
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic runtime manifest to load, got error: %v", err)
	}

	cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	cleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupPluginMenuRowsHard(t, ctx, pluginID)
		cleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected runtime plugin manifest sync to succeed, got error: %v", err)
	}
	if err = service.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin install to succeed, got error: %v", err)
	}

	menu, err := queryMenuByKey(ctx, buildDynamicRoutePermissionMenuKey(pluginID, "plugin-dynamic-route-permission:review:view"))
	if err != nil {
		t.Fatalf("expected synthetic permission menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatal("expected synthetic permission menu to be created")
	}
	if menu.Type != pluginMenuTypeButton.String() || menu.Visible != 0 {
		t.Fatalf("expected synthetic permission menu to be hidden button, got %#v", menu)
	}
}

func createTestRuntimeStorageArtifactWithMenus(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	menus []*pluginMenuSpec,
	installSQLAssets []*pluginDynamicArtifactSQLAsset,
	uninstallSQLAssets []*pluginDynamicArtifactSQLAsset,
) string {
	t.Helper()

	repoRoot, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	storageDir := filepath.Join(repoRoot, "temp", "output")
	if err = os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, buildPluginDynamicArtifactFileName(pluginID))
	t.Cleanup(func() {
		_ = os.Remove(artifactPath)
	})

	writeRuntimeWasmArtifact(
		t,
		artifactPath,
		&pluginDynamicArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    pluginTypeDynamic.String(),
			Menus:   menus,
		},
		&pluginDynamicArtifactMetadata{
			RuntimeKind:        pluginDynamicKindWasm.String(),
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(defaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		defaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
	return artifactPath
}

func cleanupPluginMenuRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	menus, err := New().listPluginMenusByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin menu cleanup query to succeed, got error: %v", err)
	}
	if len(menus) == 0 {
		return
	}

	menuIDs := make([]interface{}, 0, len(menus))
	menuKeys := make([]string, 0, len(menus))
	for _, item := range menus {
		if item == nil {
			continue
		}
		menuIDs = append(menuIDs, item.Id)
		menuKeys = append(menuKeys, item.MenuKey)
	}

	if len(menuIDs) > 0 {
		_, _ = dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).
			Delete()
	}
	if len(menuKeys) > 0 {
		_, _ = dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().MenuKey, menuKeys).
			Delete()
	}
}

func queryMenuByKey(ctx context.Context, menuKey string) (*entity.SysMenu, error) {
	var menu *entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{MenuKey: menuKey}).
		Scan(&menu)
	return menu, err
}
