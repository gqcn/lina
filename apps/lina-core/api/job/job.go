// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package job

import (
	"context"

	"lina-core/api/job/v1"
)

type IJobV1 interface {
	JobCreate(ctx context.Context, req *v1.JobCreateReq) (res *v1.JobCreateRes, err error)
	JobDelete(ctx context.Context, req *v1.JobDeleteReq) (res *v1.JobDeleteRes, err error)
	JobList(ctx context.Context, req *v1.JobListReq) (res *v1.JobListRes, err error)
	JobLogList(ctx context.Context, req *v1.JobLogListReq) (res *v1.JobLogListRes, err error)
	JobRun(ctx context.Context, req *v1.JobRunReq) (res *v1.JobRunRes, err error)
	JobStatus(ctx context.Context, req *v1.JobStatusReq) (res *v1.JobStatusRes, err error)
	JobUpdate(ctx context.Context, req *v1.JobUpdateReq) (res *v1.JobUpdateRes, err error)
}
