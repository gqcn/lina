package locker

import (
	"context"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/os/gtime"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

// Lock 获取分布式锁
func (s *Service) Lock(ctx context.Context, name, reason string, duration time.Duration) (instance *Instance, ok bool, err error) {
	var locker *entity.SysLocker
	err = dao.SysLocker.Ctx(ctx).Where(do.SysLocker{Name: name}).Scan(&locker)
	if err != nil {
		return nil, false, err
	}

	if locker == nil {
		result, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			Name:       name,
			Reason:     reason,
			CreateTime: gtime.Now(),
			ExpireTime: gtime.Now().Add(duration),
		}).InsertIgnore()
		if err != nil {
			return nil, false, err
		}
		insertId, err := result.LastInsertId()
		if err != nil {
			return nil, false, err
		}
		if insertId <= 0 {
			return nil, false, nil
		}
		return &Instance{Id: insertId}, true, nil
	}

	if gtime.Now().Before(locker.ExpireTime) {
		return nil, false, nil
	}

	affected, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		Reason:     reason,
		CreateTime: gtime.Now(),
		ExpireTime: gtime.Now().Add(duration),
	}).Where(do.SysLocker{Id: locker.Id}).UpdateAndGetAffected()
	if err != nil {
		return nil, false, err
	}
	if affected <= 0 {
		return nil, false, nil
	}
	return &Instance{Id: int64(locker.Id)}, true, nil
}

// LockFunc 获取锁并执行函数，执行完自动释放
func (s *Service) LockFunc(ctx context.Context, name, reason string, duration time.Duration, f func() error) (ok bool, err error) {
	instance, ok, err := s.Lock(ctx, name, reason, duration)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	defer func() {
		_ = instance.Unlock(ctx)
	}()
	if err = f(); err != nil {
		return true, err
	}
	return true, nil
}
