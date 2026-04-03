// Package election 提供基于分布式锁的领导选举服务。
// 负责节点标识生成、选主循环、租约续期和故障转移。
package election

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"lina-core/internal/service/config"
	"lina-core/internal/service/locker"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gipv4"
)

// lockName is the distributed lock name used for leader election.
const lockName = "leader-election"

// Service manages the leader election process.
type Service struct {
	locker      *locker.Service        // Distributed lock service
	cfg         *config.ElectionConfig // Election configuration (lease, renewInterval)
	holder      string                 // Current node identifier
	isLeader    atomic.Bool            // Whether current node is the leader
	instance    *locker.Instance       // Current lock instance (non-nil when leader)
	leaseMgr    *locker.LeaseManager   // Lease renewal manager (non-nil when leader)
	stopChan    chan struct{}
	stoppedChan chan struct{}
	once        sync.Once // Ensure Start is only called once
	stoppedOnce sync.Once // Ensure stoppedChan is closed only once
}

// New creates and returns a new election Service.
// The node identifier is automatically generated from hostname or intranet IP.
func New(lockerSvc *locker.Service, cfg *config.ElectionConfig) *Service {
	s := &Service{
		locker:   lockerSvc,
		cfg:      cfg,
		holder:   generateNodeIdentifier(),
		stopChan: make(chan struct{}),
	}
	// Pre-close stoppedChan so Stop() is safe before Start() is called
	s.stoppedChan = make(chan struct{})
	close(s.stoppedChan)
	return s
}

// Start begins the leader election process.
// It is safe to call only once; subsequent calls are no-ops.
func (s *Service) Start(ctx context.Context) {
	s.once.Do(func() {
		s.stoppedChan = make(chan struct{})
		go s.run(ctx)
	})
}

// Stop stops the leader election and releases leadership if held.
func (s *Service) Stop(ctx context.Context) {
	select {
	case <-s.stopChan:
		// Already stopped
	default:
		close(s.stopChan)
	}
	<-s.stoppedChan
}

// IsLeader returns whether the current node is the leader.
func (s *Service) IsLeader() bool {
	return s.isLeader.Load()
}

// Holder returns the current node identifier.
func (s *Service) Holder() string {
	return s.holder
}

// run is the main election loop.
func (s *Service) run(ctx context.Context) {
	defer s.stoppedOnce.Do(func() { close(s.stoppedChan) })

	s.tryAcquire(ctx)

	retryTicker := time.NewTicker(s.cfg.RenewInterval)
	defer retryTicker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.stepDown(ctx)
			g.Log().Infof(ctx, "[election] leader election stopped")
			return
		case <-s.leaseStoppedChan():
			g.Log().Warningf(ctx, "[election] lease renewal stopped, attempting to re-acquire")
			s.instance = nil
			s.leaseMgr = nil
			s.isLeader.Store(false)
			retryTicker.Reset(s.cfg.RenewInterval)
			s.tryAcquire(ctx)
		case <-retryTicker.C:
			if !s.isLeader.Load() {
				s.tryAcquire(ctx)
			}
		}
	}
}

// tryAcquire attempts to acquire the leader lock.
func (s *Service) tryAcquire(ctx context.Context) {
	instance, ok, err := s.locker.Lock(ctx, lockName, s.holder, "leader election", s.cfg.Lease)
	if err != nil {
		g.Log().Warningf(ctx, "[election] failed to acquire lock: %v", err)
		s.isLeader.Store(false)
		return
	}

	if ok {
		s.instance = instance
		s.isLeader.Store(true)
		g.Log().Infof(ctx, "[election] became leader (holder: %s)", s.holder)

		s.leaseMgr = locker.NewLeaseManager(instance, s.cfg.RenewInterval)
		s.leaseMgr.Start(ctx)
	} else {
		s.isLeader.Store(false)
		g.Log().Debugf(ctx, "[election] not leader, waiting for lease expiry")
	}
}

// stepDown releases leadership and stops lease renewal.
func (s *Service) stepDown(ctx context.Context) {
	if s.leaseMgr != nil {
		s.leaseMgr.Stop()
		s.leaseMgr = nil
	}
	if s.instance != nil {
		if err := s.instance.Unlock(ctx); err != nil {
			g.Log().Warningf(ctx, "[election] failed to release lock: %v", err)
		}
		s.instance = nil
	}
	s.isLeader.Store(false)
	g.Log().Infof(ctx, "[election] stepped down from leadership")
}

// leaseStoppedChan returns the channel that closes when lease renewal stops.
// Returns nil (blocks forever in select) when not the leader.
func (s *Service) leaseStoppedChan() <-chan struct{} {
	if s.leaseMgr != nil {
		return s.leaseMgr.StoppedChan()
	}
	return nil
}

// generateNodeIdentifier generates a unique node identifier using hostname or intranet IP.
func generateNodeIdentifier() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname, _ = gipv4.GetIntranetIp()
	}
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}
