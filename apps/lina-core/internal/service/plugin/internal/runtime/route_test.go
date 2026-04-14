// This file covers dynamic-route-specific session validation behaviors that
// are easy to regress during runtime auth changes.

package runtime

import (
	"context"
	`encoding/json`
	"fmt"
	`net/http`
	`path/filepath`
	"testing"
	"time"

	`lina-core/internal/service/plugin/internal/testutil`
	`lina-core/pkg/pluginbridge`

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

func TestTouchDynamicRouteSessionKeepsExistingSessionWhenTimestampDoesNotChange(t *testing.T) {
	var (
		ctx     = context.Background()
		service = &Service{}
		tokenID = fmt.Sprintf("plugin-dynamic-route-session-%d", time.Now().UnixNano())
	)

	_, _ = dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Delete()
	defer func() {
		_, _ = dao.SysOnlineSession.Ctx(ctx).
			Where(do.SysOnlineSession{TokenId: tokenID}).
			Delete()
	}()

	currentSecond := waitForFreshSecond(t)
	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        tokenID,
		UserId:         1,
		Username:       "admin",
		DeptName:       "系统管理",
		Ip:             "127.0.0.1",
		Browser:        "go-test",
		Os:             "darwin",
		LoginTime:      currentSecond,
		LastActiveTime: currentSecond,
	}).Insert()
	if err != nil {
		t.Fatalf("expected test session insert to succeed, got error: %v", err)
	}

	exists, err := service.touchDynamicRouteSession(ctx, tokenID)
	if err != nil {
		t.Fatalf("expected first session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected first session touch to keep the session active")
	}

	// Touch the same session again within the same second to emulate the dynamic
	// route request arriving immediately after login or another protected request.
	exists, err = service.touchDynamicRouteSession(ctx, tokenID)
	if err != nil {
		t.Fatalf("expected second session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected existing session to remain active when DATETIME precision keeps the same second")
	}
}

func waitForFreshSecond(t *testing.T) *gtime.Time {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		now := time.Now()
		if now.Nanosecond() < int((100 * time.Millisecond).Nanoseconds()) {
			return gtime.NewFromTime(now.Truncate(time.Second))
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("failed to align test to the beginning of a second")
	return nil
}

func TestMatchDynamicRoutePathSupportsParams(t *testing.T) {
	params, ok := runtime.MatchDynamicRoutePath("/records/{id}/detail", "/records/42/detail")
	if !ok {
		t.Fatal("expected dynamic path match to succeed")
	}
	if params["id"] != "42" {
		t.Fatalf("expected path param id=42, got %#v", params)
	}
}

func TestBuildDynamicRouteOperLogMetadataMapsRouteGovernance(t *testing.T) {
	metadata := runtime.BuildDynamicRouteOperLogMetadata(&runtime.DynamicRouteRuntimeState{
		Match: &runtime.DynamicRouteMatch{
			Route: &pluginbridge.RouteContract{
				Tags:    []string{"plugin-review", "dynamic"},
				Summary: "Review summary",
				OperLog: "other",
			},
		},
	})
	if metadata == nil {
		t.Fatal("expected dynamic route operlog metadata to be built")
	}
	if metadata.Title != "plugin-review,dynamic" {
		t.Fatalf("expected title to join route tags, got %q", metadata.Title)
	}
	if metadata.Summary != "Review summary" {
		t.Fatalf("expected summary to be preserved, got %q", metadata.Summary)
	}
	if metadata.OperLogTag != "other" {
		t.Fatalf("expected operlog tag other, got %q", metadata.OperLogTag)
	}
}

func TestExecuteDynamicWasmBridgeReturnsGuestResponse(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	artifactPath := filepath.Join(repoRoot, "temp", "output", runtime.BuildArtifactFileName("plugin-demo-dynamic"))
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected bundled runtime artifact to load, got error: %v", err)
	}

	response, err := services.Runtime.ExecuteDynamicRoute(context.Background(), manifest, &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: "plugin-demo-dynamic",
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
		t.Fatalf("expected dynamic wasm execution to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != http.StatusOK {
		t.Fatalf("expected guest bridge response 200, got %#v", response)
	}
	if string(response.Body) == "" {
		t.Fatal("expected guest bridge response body to be non-empty")
	}
	if got := response.Headers["X-Lina-Plugin-Bridge"]; len(got) != 1 || got[0] != "plugin-demo-dynamic" {
		t.Fatalf("expected guest bridge header to be preserved, got %#v", response.Headers)
	}
	if got := response.Headers["X-Lina-Plugin-Middleware"]; len(got) != 1 || got[0] != "backend-summary" {
		t.Fatalf("expected guest-local middleware header to be preserved, got %#v", response.Headers)
	}

	payload := map[string]interface{}{}
	if err = json.Unmarshal(response.Body, &payload); err != nil {
		t.Fatalf("expected guest response body to be valid json, got error: %v", err)
	}
	if payload["pluginId"] != "plugin-demo-dynamic" {
		t.Fatalf("expected guest payload pluginId to be preserved, got %#v", payload)
	}
	if payload["authenticated"] != true {
		t.Fatalf("expected guest payload authenticated=true, got %#v", payload)
	}
}
