package sysinfo

import (
	"context"

	"lina-core/api/sysinfo/v1"
)

func (c *ControllerV1) GetInfo(ctx context.Context, req *v1.GetInfoReq) (res *v1.GetInfoRes, err error) {
	info, err := c.sysInfoSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetInfoRes{
		GoVersion:   info.GoVersion,
		GfVersion:   info.GfVersion,
		Os:          info.Os,
		Arch:        info.Arch,
		DbVersion:   info.DbVersion,
		StartTime:   info.StartTime,
		RunDuration: info.RunDuration,
	}, nil
}
