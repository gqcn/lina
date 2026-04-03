package election

import (
	"context"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
)

// testCfg is the default election config used in tests.
var testCfg = &config.ElectionConfig{
	Lease:         30 * time.Second,
	RenewInterval: 1 * time.Second,
}

// newTestService creates an election service for testing.
func newTestService() *Service {
	return New(locker.New(), testCfg)
}

// cleanupLock removes the election lock after test.
func cleanupLock() {
	_, _ = g.DB().Model("sys_locker").Where("name", lockName).Delete()
}

func TestElection_New(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestService()

		t.AssertNE(svc, nil)
		t.AssertNE(svc.Holder(), "")
		t.Assert(svc.IsLeader(), false)
	})
}

func TestElection_StartAndBecomeLeader(t *testing.T) {
	var (
		svc = newTestService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		count, err := g.DB().Model("sys_locker").Where("name", lockName).Count()
		t.AssertNil(err)
		t.Assert(count, 1)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

func TestElection_AlreadyLeader(t *testing.T) {
	var (
		svc = newTestService()
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(30 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), false)

		svc.Stop(ctx)
	})

	cleanupLock()
}

func TestElection_TakeOverExpiredLock(t *testing.T) {
	var (
		svc = newTestService()
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(-10 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		var row struct{ Holder string }
		err = g.DB().Model("sys_locker").Where("name", lockName).Scan(&row)
		t.AssertNil(err)
		t.Assert(row.Holder, svc.Holder())

		svc.Stop(ctx)
	})

	cleanupLock()
}

func TestElection_StepDown(t *testing.T) {
	var (
		svc = newTestService()
		ctx = context.Background()
	)

	cleanupLock()

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(200 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		svc.Stop(ctx)

		t.Assert(svc.IsLeader(), false)
	})

	cleanupLock()
}

func TestElection_StopWithoutStart(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		svc := newTestService()
		svc.Stop(context.Background())
	})
}

func TestElection_NonLeaderRetry(t *testing.T) {
	var (
		retryCfg = &config.ElectionConfig{
			Lease:         30 * time.Second,
			RenewInterval: 200 * time.Millisecond,
		}
		svc = New(locker.New(), retryCfg)
		ctx = context.Background()
	)

	cleanupLock()

	_, err := g.DB().Model("sys_locker").Data(g.Map{
		"name":        lockName,
		"reason":      "election",
		"holder":      "other-node",
		"expire_time": gtime.Now().Add(-5 * time.Second),
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}

	gtest.C(t, func(t *gtest.T) {
		svc.Start(ctx)

		time.Sleep(500 * time.Millisecond)

		t.Assert(svc.IsLeader(), true)

		svc.Stop(ctx)
	})

	cleanupLock()
}
