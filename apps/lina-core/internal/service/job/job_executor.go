package job

import (
	"context"
	"fmt"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
)

var systemJobHandlers = map[string]func(context.Context) error{}

func RegisterSystemHandler(name string, handler func(context.Context) error) {
	systemJobHandlers[name] = handler
}

func (s *Service) Execute(ctx context.Context, job *entity.SysJob) error {
	if job.Status == 0 {
		return nil
	}

	if job.MaxTimes > 0 && job.ExecTimes >= job.MaxTimes {
		return nil
	}

	lockName := fmt.Sprintf("job:%d", job.Id)
	if job.Singleton == 1 {
		ok, err := s.lockerSvc.LockFunc(ctx, lockName, "job execution", time.Hour, func() error {
			return s.executeWithLog(ctx, job)
		})
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return nil
	}

	return s.executeWithLog(ctx, job)
}

func (s *Service) executeWithLog(ctx context.Context, job *entity.SysJob) error {
	startTime := gtime.Now()
	var logId int64
	var execErr error

	result, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:      job.Id,
		JobName:    job.Name,
		JobGroup:   job.Group,
		Command:    job.Command,
		Status:     0,
		StartTime:  startTime,
		CreateTime: startTime,
	}).Insert()
	if err == nil {
		logId, _ = result.LastInsertId()
	}

	execErr = s.executeCommand(ctx, job.Command)

	endTime := gtime.Now()
	duration := endTime.Sub(startTime).Milliseconds()
	status := 1
	var errorMsg string
	if execErr != nil {
		status = 0
		errorMsg = execErr.Error()
	}

	if logId > 0 {
		_, _ = dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
			Status:   status,
			EndTime:  endTime,
			Duration: int(duration),
			ErrorMsg: errorMsg,
		}).Where(do.SysJobLog{Id: logId}).Update()
	}

	_, _ = dao.SysJob.Ctx(ctx).Data(do.SysJob{
		ExecTimes: job.ExecTimes + 1,
	}).Where(do.SysJob{Id: job.Id}).Update()

	if job.MaxTimes > 0 && job.ExecTimes+1 >= job.MaxTimes {
		_, _ = dao.SysJob.Ctx(ctx).Data(do.SysJob{Status: 0}).Where(do.SysJob{Id: job.Id}).Update()
	}

	return execErr
}

func (s *Service) executeCommand(ctx context.Context, command string) error {
	if gstr.HasPrefix(command, "<") && gstr.HasSuffix(command, ">") {
		handlerName := gstr.Trim(command, "<>")
		handler, ok := systemJobHandlers[handlerName]
		if !ok {
			return fmt.Errorf("system handler not found: %s", handlerName)
		}
		return handler(ctx)
	}

	result, err := gproc.ShellExec(ctx, command)
	if err != nil {
		return err
	}
	if result != "" && gstr.Contains(result, "error") {
		return fmt.Errorf("command output: %s", result)
	}
	return nil
}
