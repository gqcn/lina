package locker

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"

	"github.com/gogf/gf/v2/os/gtime"
)

type Instance struct {
	Id int64
}

// Unlock 释放锁
func (i *Instance) Unlock(ctx context.Context) error {
	_, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: gtime.Now(),
	}).Where(do.SysLocker{Id: i.Id}).Update()
	return err
}
