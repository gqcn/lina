# v0.8.0 定时任务功能 - 技术设计

## 架构设计

### 系统架构
```
┌─────────────────────────────────────────────────────────────┐
│                         前端层                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  任务列表页   │  │  任务表单     │  │  执行日志页   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                         后端层                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Controller (HTTP 接口层)                 │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  Service (业务逻辑层)                  │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │  │
│  │  │   job    │  │  locker  │  │   cron   │           │  │
│  │  └──────────┘  └──────────┘  └──────────┘           │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  DAO (数据访问层)                      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       数据库层                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │
│  │ sys_job  │  │sys_job_log│ │sys_locker│                 │
│  └──────────┘  └──────────┘  └──────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

## 数据库设计

### sys_job (定时任务表)
```sql
CREATE TABLE `sys_job` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `name` varchar(64) NOT NULL COMMENT '任务名称',
  `group` varchar(64) NOT NULL DEFAULT 'default' COMMENT '任务分组',
  `command` varchar(500) NOT NULL COMMENT '执行指令',
  `cron_expr` varchar(255) NOT NULL COMMENT 'Cron表达式',
  `description` varchar(500) DEFAULT NULL COMMENT '任务描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=禁用',
  `singleton` tinyint NOT NULL DEFAULT 1 COMMENT '执行模式：1=单例 0=并行',
  `max_times` int NOT NULL DEFAULT 0 COMMENT '最大执行次数，0表示无限制',
  `exec_times` int NOT NULL DEFAULT 0 COMMENT '已执行次数',
  `is_system` tinyint NOT NULL DEFAULT 0 COMMENT '是否系统任务：1=是 0=否',
  `create_by` varchar(64) DEFAULT NULL COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT NULL COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_group` (`group`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';
```

### sys_job_log (任务执行日志表)
```sql
CREATE TABLE `sys_job_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `job_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `job_name` varchar(64) NOT NULL COMMENT '任务名称',
  `job_group` varchar(64) NOT NULL COMMENT '任务分组',
  `command` varchar(500) NOT NULL COMMENT '执行指令',
  `status` tinyint NOT NULL COMMENT '执行状态：1=成功 0=失败',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `duration` int DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `error_msg` text COMMENT '错误信息',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表';
```

### sys_locker (分布式锁表)
```sql
CREATE TABLE `sys_locker` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '锁ID',
  `name` varchar(255) NOT NULL COMMENT '锁名称',
  `reason` varchar(500) DEFAULT NULL COMMENT '锁定原因',
  `create_time` datetime NOT NULL COMMENT '创建时间',
  `expire_time` datetime NOT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';
```

## 后端设计

### 目录结构
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
│   │   └── job_v1_*.go      # 控制器(自动生成)
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
│   ├── dao/                 # DAO层(自动生成)
│   └── model/
│       ├── entity/          # 实体(自动生成)
│       └── do/              # DO对象(自动生成)
└── manifest/sql/
    └── v0.8.0.sql           # 数据库变更脚本
```

### 核心服务设计

#### 1. locker 服务 (分布式锁)
```go
type Service struct{}

// Lock 获取分布式锁
func (s *Service) Lock(ctx context.Context, name, reason string, duration time.Duration) (instance *Instance, ok bool, err error)

// LockFunc 获取锁并执行函数，自动释放
func (s *Service) LockFunc(ctx context.Context, name, reason string, duration time.Duration, f func() error) (ok bool, err error)

type Instance struct {
    Id int64
}

// Unlock 释放锁
func (i *Instance) Unlock(ctx context.Context) error
```

#### 2. job 服务 (任务管理)
```go
type Service struct {
    lockerSvc *locker.Service
}

// List 查询任务列表
func (s *Service) List(ctx context.Context, in *model.JobListInput) (*model.JobListOutput, error)

// Create 创建任务
func (s *Service) Create(ctx context.Context, in *model.JobCreateInput) error

// Update 更新任务
func (s *Service) Update(ctx context.Context, in *model.JobUpdateInput) error

// Delete 删除任务
func (s *Service) Delete(ctx context.Context, ids []int64) error

// UpdateStatus 更新任务状态
func (s *Service) UpdateStatus(ctx context.Context, id int64, status int) error

// Run 手动执行任务
func (s *Service) Run(ctx context.Context, id int64) error

// Execute 执行任务(内部方法)
func (s *Service) Execute(ctx context.Context, job *entity.SysJob) error
```

#### 3. cron 服务 (任务调度)
```go
type Service struct {
    jobSvc *job.Service
}

// Start 启动所有定时任务
func (s *Service) Start(ctx context.Context)

// RegisterJob 注册任务到调度器
func (s *Service) RegisterJob(ctx context.Context, job *entity.SysJob) error

// UnregisterJob 从调度器移除任务
func (s *Service) UnregisterJob(ctx context.Context, jobId int64) error

// ReloadJobs 重新加载所有任务
func (s *Service) ReloadJobs(ctx context.Context) error
```

### 任务执行流程

```
1. gcron 触发任务执行
   ↓
2. 检查任务状态(是否启用)
   ↓
3. 检查执行次数(是否达到上限)
   ↓
4. 如果是单例模式
   ├─ 尝试获取分布式锁(job:{id})
   ├─ 获取失败 → 跳过本次执行
   └─ 获取成功 → 继续
   ↓
5. 记录执行开始(插入日志记录)
   ↓
6. 执行任务指令
   ├─ 系统任务: 调用注册的Go函数
   └─ 自定义任务: 执行shell命令
   ↓
7. 更新执行日志(结束时间、耗时、状态、错误信息)
   ↓
8. 更新已执行次数(exec_times + 1)
   ↓
9. 检查是否达到最大次数
   ├─ 是 → 禁用任务
   └─ 否 → 继续
   ↓
10. 释放分布式锁(如果是单例模式)
```

### 系统任务注册

在 `service/cron/cron_job.go` 中维护系统任务映射：

```go
var systemJobHandlers = map[string]func(context.Context) error{
    "session.Cleanup":    sessionCleanupHandler,
    "servermon.Collect":  servermonCollectHandler,
}

func sessionCleanupHandler(ctx context.Context) error {
    return session.New().Cleanup(ctx)
}

func servermonCollectHandler(ctx context.Context) error {
    return servermon.New().Collect(ctx)
}
```

## 前端设计

### 路由配置
```typescript
// src/router/routes/modules/monitor.ts
{
  path: 'job',
  name: 'MonitorJob',
  component: () => import('#/views/monitor/job/index.vue'),
  meta: { title: '定时任务' }
}
```

### 页面结构

#### 1. 任务列表页 (index.vue)
- 搜索区域: 任务名称、分组、状态
- 工具栏: 新增按钮、刷新按钮
- 表格: VXE-Grid
  - 列: 任务名称、分组、指令、Cron表达式、状态、执行策略、执行次数、操作
  - 操作: 编辑、删除、启用/禁用、执行一次、查看日志
- 表单弹窗: 新增/编辑任务

#### 2. 执行日志页 (log.vue)
- 搜索区域: 任务名称、执行状态、时间范围
- 表格: VXE-Grid
  - 列: 任务名称、分组、开始时间、结束时间、耗时、状态、操作
  - 操作: 查看详情
- 详情抽屉: 显示完整的执行信息和错误信息

### API 接口

```typescript
// src/api/monitor/job.ts
export const jobApi = {
  list: (params) => requestClient.get('/job/list', { params }),
  create: (data) => requestClient.post('/job/create', data),
  update: (data) => requestClient.put('/job/update', data),
  delete: (ids) => requestClient.delete('/job/delete', { data: { ids } }),
  updateStatus: (id, status) => requestClient.put('/job/status', { id, status }),
  run: (id) => requestClient.post('/job/run', { id }),
  logList: (params) => requestClient.get('/job/log/list', { params }),
}
```

## 初始化数据

### 系统任务
```sql
INSERT INTO `sys_job` VALUES
(1, '会话清理', 'system', '<session.Cleanup>', '0 0 * * * *', '清理过期的用户会话', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL),
(2, '服务器监控', 'system', '<servermon.Collect>', '0 * * * * *', '采集服务器性能指标', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL);
```

### 字典数据
```sql
-- 任务状态
INSERT INTO `sys_dict_type` VALUES (NULL, '任务状态', 'sys_job_status', 1, 'admin', NOW(), NULL, NULL, '定时任务状态');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_status'), '启用', '1', 1, 'success', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_status'), '禁用', '0', 2, 'danger', 'admin', NOW(), NULL, NULL, NULL);

-- 执行模式
INSERT INTO `sys_dict_type` VALUES (NULL, '任务执行模式', 'sys_job_singleton', 1, 'admin', NOW(), NULL, NULL, '定时任务执行模式');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_singleton'), '单例执行', '1', 1, 'default', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_singleton'), '并行执行', '0', 2, 'default', 'admin', NOW(), NULL, NULL, NULL);

-- 执行状态
INSERT INTO `sys_dict_type` VALUES (NULL, '任务执行状态', 'sys_job_log_status', 1, 'admin', NOW(), NULL, NULL, '任务执行日志状态');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_log_status'), '成功', '1', 1, 'success', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_log_status'), '失败', '0', 2, 'danger', 'admin', NOW(), NULL, NULL, NULL);
```

## 技术要点

### 1. 分布式锁实现
- 使用数据库唯一索引保证锁的互斥性
- 通过 `expire_time` 实现锁的自动过期
- 使用 `InsertIgnore` 和 `UpdateAndGetAffected` 保证原子性

### 2. 任务动态注册
- 启动时加载所有启用的任务
- 任务增删改时动态更新 gcron 调度器
- 使用任务ID作为 gcron 的任务名称

### 3. 执行次数控制
- 每次执行后更新 `exec_times`
- 达到 `max_times` 后自动禁用任务
- 使用事务保证计数的准确性

### 4. 错误处理
- 任务执行失败记录详细错误信息
- 不影响后续任务的执行
- 分布式锁获取失败不记录日志(正常跳过)

## 性能优化

1. **索引优化**: 在 `group`、`status`、`job_id`、`start_time` 等字段上建立索引
2. **日志清理**: 定期清理过期的执行日志(可作为系统任务)
3. **锁超时**: 分布式锁设置合理的超时时间,避免死锁
4. **并发控制**: 限制并行任务的最大并发数

## 安全考虑

1. **权限控制**: 只有管理员可以管理定时任务
2. **命令校验**: 自定义任务的命令需要进行安全校验
3. **系统任务保护**: 系统任务不可删除,指令不可修改
4. **日志脱敏**: 执行日志中的敏感信息需要脱敏处理
