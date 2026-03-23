# 后端规范

## 目录结构

```
apps/lina-core/
├── api/job/v1/
│   ├── job_list.go          # 任务列表查询
│   ├── job_create.go        # 创建任务
│   ├── job_update.go        # 更新任务
│   ├── job_delete.go        # 删除任务
│   ├── job_status.go        # 启用/禁用任务
│   ├── job_run.go           # 手动执行任务
│   └── job_log_list.go      # 执行日志列表
├── internal/
│   ├── controller/job/
│   │   └── job_v1_*.go      # 控制器(gf gen ctrl 生成)
│   ├── service/
│   │   ├── job/
│   │   │   ├── job.go       # 任务管理服务
│   │   │   ├── job_executor.go  # 任务执行器
│   │   │   └── job_log.go   # 日志服务
│   │   ├── locker/
│   │   │   ├── locker.go    # 分布式锁服务
│   │   │   └── locker_instance.go
│   │   └── cron/
│   │       ├── cron.go      # 定时任务调度服务
│   │       └── cron_job.go  # 动态任务注册
│   ├── dao/                 # DAO层(gf gen dao 生成)
│   └── model/
│       ├── entity/          # 实体(gf gen dao 生成)
│       └── do/              # DO对象(gf gen dao 生成)
```

## Service 层设计

### 1. locker 服务 (service/locker/)

#### locker.go
```go
package locker

import (
    "context"
    "time"
)

type Service struct{}

func New() *Service {
    return &Service{}
}

// Lock 获取分布式锁
// name: 锁名称(唯一标识)
// reason: 锁定原因
// duration: 锁持续时间
// 返回: 锁实例, 是否获取成功, 错误
func (s *Service) Lock(ctx context.Context, name, reason string, duration time.Duration) (instance *Instance, ok bool, err error)

// LockFunc 获取锁并执行函数，执行完自动释放
func (s *Service) LockFunc(ctx context.Context, name, reason string, duration time.Duration, f func() error) (ok bool, err error)
```

**实现逻辑**:
1. 查询锁记录 (WHERE name = ?)
2. 如果不存在，InsertIgnore 创建锁记录
3. 如果存在但未过期，返回 ok=false
4. 如果存在且已过期，UpdateAndGetAffected 更新锁记录
5. 根据影响行数判断是否获取成功

#### locker_instance.go
```go
package locker

import "context"

type Instance struct {
    Id int64
}

// Unlock 释放锁(更新 expire_time 为当前时间)
func (i *Instance) Unlock(ctx context.Context) error
```

### 2. job 服务 (service/job/)

#### job.go
```go
package job

import (
    "context"
    "lina-core/internal/service/locker"
)

type Service struct {
    lockerSvc *locker.Service
}

func New() *Service {
    return &Service{
        lockerSvc: locker.New(),
    }
}

// List 查询任务列表
func (s *Service) List(ctx context.Context, in *model.JobListInput) (*model.JobListOutput, error)

// Create 创建任务
func (s *Service) Create(ctx context.Context, in *model.JobCreateInput) error

// Update 更新任务
func (s *Service) Update(ctx context.Context, in *model.JobUpdateInput) error

// Delete 删除任务(系统任务不可删除)
func (s *Service) Delete(ctx context.Context, ids []int64) error

// UpdateStatus 更新任务状态
func (s *Service) UpdateStatus(ctx context.Context, id int64, status int) error

// Run 手动执行任务
func (s *Service) Run(ctx context.Context, id int64) error
```

#### job_executor.go
```go
package job

import (
    "context"
    "time"
)

// Execute 执行任务(内部方法)
// 流程:
// 1. 检查任务状态
// 2. 检查执行次数
// 3. 单例模式获取分布式锁
// 4. 记录执行开始
// 5. 执行任务指令
// 6. 记录执行结果
// 7. 更新执行次数
// 8. 释放锁
func (s *Service) Execute(ctx context.Context, job *entity.SysJob) error

// executeCommand 执行任务指令
// 系统任务: 调用注册的Go函数
// 自定义任务: 执行shell命令
func (s *Service) executeCommand(ctx context.Context, command string) error
```

**执行流程**:
```
1. 检查 job.Status (禁用则跳过)
2. 检查 job.MaxTimes (达到上限则跳过)
3. 如果 job.Singleton == 1:
   - 尝试获取锁: locker.Lock(ctx, "job:"+jobId, "job execution", 1小时)
   - 获取失败则跳过本次执行
4. 创建日志记录 (status=执行中, start_time=now)
5. 执行指令:
   - 系统任务: 从 systemJobHandlers 映射中查找函数并调用
   - 自定义任务: gproc.ShellExec(command)
6. 更新日志记录 (end_time, duration, status, error_msg)
7. 更新 job.ExecTimes += 1
8. 如果 job.MaxTimes > 0 && job.ExecTimes >= job.MaxTimes:
   - 更新 job.Status = 0 (禁用)
9. 释放锁 (如果是单例模式)
```

#### job_log.go
```go
package job

import "context"

// LogList 查询执行日志列表
func (s *Service) LogList(ctx context.Context, in *model.JobLogListInput) (*model.JobLogListOutput, error)
```

### 3. cron 服务 (service/cron/)

#### cron.go
```go
package cron

import (
    "context"
    "lina-core/internal/service/job"
)

type Service struct {
    jobSvc *job.Service
}

func New() *Service {
    return &Service{
        jobSvc: job.New(),
    }
}

// Start 启动所有定时任务
func (s *Service) Start(ctx context.Context) {
    s.startSessionCleanup(ctx)
    s.startServerMonitor(ctx)
    s.startDynamicJobs(ctx)  // 新增: 启动数据库中的动态任务
}
```

#### cron_job.go
```go
package cron

import (
    "context"
    "github.com/gogf/gf/v2/os/gcron"
)

// 系统任务处理函数映射
var systemJobHandlers = map[string]func(context.Context) error{
    "session.Cleanup":    sessionCleanupHandler,
    "servermon.Collect":  servermonCollectHandler,
}

// startDynamicJobs 启动数据库中的所有启用任务
func (s *Service) startDynamicJobs(ctx context.Context)

// RegisterJob 注册任务到调度器
func (s *Service) RegisterJob(ctx context.Context, job *entity.SysJob) error

// UnregisterJob 从调度器移除任务
func (s *Service) UnregisterJob(ctx context.Context, jobId int64) error

// ReloadJobs 重新加载所有任务
func (s *Service) ReloadJobs(ctx context.Context) error
```

**实现要点**:
- 使用 `gcron.AddSingleton` 注册任务
- 任务名称使用 "job:{id}" 格式
- 任务回调函数调用 `jobSvc.Execute(ctx, job)`
- 任务增删改时调用 `ReloadJobs` 重新加载

## API 层设计

### 接口定义示例 (api/job/v1/job_list.go)

```go
package v1

import "github.com/gogf/gf/v2/frame/g"

type JobListReq struct {
    g.Meta   `path:"/job/list" method:"get" tags:"定时任务" summary:"查询任务列表" dc:"分页查询定时任务列表，支持按任务名称、分组、状态筛选"`
    Name     string `json:"name" dc:"任务名称，支持模糊查询" eg:"会话清理"`
    Group    string `json:"group" dc:"任务分组" eg:"system"`
    Status   *int   `json:"status" dc:"任务状态：1=启用 0=禁用，不传则查询全部" eg:"1"`
    Page     int    `json:"page" v:"required|min:1" dc:"页码" eg:"1"`
    PageSize int    `json:"pageSize" v:"required|min:1|max:100" dc:"每页数量" eg:"10"`
}

type JobListRes struct {
    Items []*JobItem `json:"items" dc:"任务列表"`
    Total int        `json:"total" dc:"总数"`
}

type JobItem struct {
    Id          uint64 `json:"id" dc:"任务ID"`
    Name        string `json:"name" dc:"任务名称"`
    Group       string `json:"group" dc:"任务分组"`
    Command     string `json:"command" dc:"执行指令"`
    CronExpr    string `json:"cronExpr" dc:"Cron表达式"`
    Description string `json:"description" dc:"任务描述"`
    Status      int    `json:"status" dc:"状态：1=启用 0=禁用"`
    Singleton   int    `json:"singleton" dc:"执行模式：1=单例 0=并行"`
    MaxTimes    int    `json:"maxTimes" dc:"最大执行次数"`
    ExecTimes   int    `json:"execTimes" dc:"已执行次数"`
    IsSystem    int    `json:"isSystem" dc:"是否系统任务：1=是 0=否"`
    CreateBy    string `json:"createBy" dc:"创建者"`
    CreateTime  string `json:"createTime" dc:"创建时间"`
    UpdateBy    string `json:"updateBy" dc:"更新者"`
    UpdateTime  string `json:"updateTime" dc:"更新时间"`
    Remark      string `json:"remark" dc:"备注"`
}
```

## 业务规则

### 1. 系统任务保护
- 系统任务 (is_system=1) 不可删除
- 系统任务的 command 字段不可修改
- 系统任务的 command 显示为 `<函数名>` 格式

### 2. 执行次数控制
- max_times=0 表示无限制
- 每次执行后 exec_times += 1
- 当 exec_times >= max_times 时自动禁用任务

### 3. 单例执行
- singleton=1 时使用分布式锁
- 锁名称: "job:{id}"
- 锁超时时间: 1小时 (防止任务执行时间过长)
- 获取锁失败时跳过本次执行，不记录日志

### 4. 任务状态变更
- 启用任务时自动注册到 gcron
- 禁用任务时从 gcron 移除
- 删除任务时从 gcron 移除

## 错误处理

1. **任务执行失败**: 记录错误信息到日志表，不影响后续执行
2. **锁获取失败**: 跳过本次执行，不记录日志
3. **系统任务删除**: 返回错误 "系统任务不可删除"
4. **Cron表达式错误**: 创建/更新时校验，返回错误提示
