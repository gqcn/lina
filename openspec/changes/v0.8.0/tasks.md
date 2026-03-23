# v0.8.0 定时任务功能 - 实现任务清单

## 数据库变更

### Task 1: 创建数据库表和初始化数据
- [x] 创建 `manifest/sql/v0.8.0.sql` 文件
- [x] 定义 `sys_job` 表结构
- [x] 定义 `sys_job_log` 表结构
- [x] 定义 `sys_locker` 表结构
- [x] 插入系统任务初始化数据
- [x] 插入字典类型和字典数据
- [x] 执行 `make init` 应用数据库变更
- [x] 执行 `make dao` 生成 DAO/DO/Entity

## 后端实现

### Task 2: 实现分布式锁服务
- [x] 创建 `internal/service/locker/locker.go`
- [x] 实现 `Lock` 法
- [x] 实现 `LockFunc` 方法
- [x] 创建 `internal/service/locker/locker_instance.go`
- [x] 实现 `Unlock` 方法

### Task 3: 定义任务管理 API 接口
- [x] 创建 `api/job/v1/job_list.go` (任务列表查询)
- [x] 创建 `api/job/v1/job_create.go` (创建任务)
- [x] 创建 `api/job/v1/job_update.go` (更新任务)
- [x] 创建 `api/job/v1/job_delete.go` (删除任务)
- [x] 创建 `api/job/v1/job_status.go` (启用/禁用任务)
- [x] 创建 `api/job/v1/job_run.go` (手动执行任务)
- [x] 创建 `api/job/v1/job_log_list.go` (执行日志列表)
- [x] 执行 `make ctrl` 生成控制器骨架

### Task 4: 实现任务管理服务
- [x] 创建 `internal/service/job/job.go`
- [x] 实现 `List` 方法 (任务列表查询)
- [x] 实现 `Create` 方法 (创建任务)
- [x] 实现 `Update` 方法 (更新任务，系统任务指令不可修改)
- [x] 实现 `Delete` 方法 (删除任务，系统任务不可删除)
- [x] 实现 `UpdateStatus` 方法 (启用/禁用任务)
- [x] 实现 `Run` 方法 (手动执行任务)

### Task 5: 实现任务执行器
- [x] 创建 `internal/service/job/job_executor.go`
- [x] 实现 `Execute` 方法 (任务执行主流程)
- [x] 实现 `executeCommand` 方法 (执行指令)
- [x] 实现单例执行逻辑 (基于分布式锁)
- [x] 实现执行次数控制逻辑
- [x] 实现执行日志记录

### Task 6: 实现任务日志服务
- [x] 创建 `internal/service/job/job_log.go`
- [x] 实现 `LogList` 方法 (日志列表查询)

### Task 7: 扩展 cron 服务支持动态任务
- [x] 创建 `internal/service/cron/cron_job.go`
- [x] 定义系统任务处理函数映射
- [x] 实现 `startDynamicJobs` 方法 (启动数据库中的任务)
- [x] 实现 `RegisterJob` 方法 (注册任务到调度器)
- [x] 实现 `UnregisterJob` 方法 (移除任务)
- [x] 实现 `ReloadJobs` 方法 (重新加载所有任务)
- [x] 修改 `cron.go` 的 `Start` 方法调用 `startDynamicJobs`

### Task 8: 实现控制器逻辑
- [x] 实现 `controller/job/job_v1_job_list.go`
- [x] 实现 `controller/job/job_v1_job_create.go`
- [x] 实现 `controller/job/job_v1_job_update.go`
- [x] 实现 `controller/job/job_v1_job_delete.go`
- [x] 实现 `controller/job/job_v1_job_status.go`
- [x] 实现 `controller/job/job_v1_job_run.go`
- [x] 实现 `controller/job/job_v1_job_log_list.go`

## 前端实现

### Task 9: 创建任务管理 API
- [x] 创建 `apps/lina-vben/apps/web-antd/src/api/monitor/job.ts`
- [x] 定义 TypeScript 接口类型
- [x] 实现所有 API 方法

### Task 10: 实现任务列表页
- [x] 创建 `apps/lina-vben/apps/web-antd/src/views/monitor/job/index.vue`
- [x] 实现搜索区域 (任务名称、分组、状态)
- [x] 实现工具栏 (新增、刷新按钮)
- [x] 实现 VXE-Grid 表格
- [x] 实现操作列按钮 (编辑、删除、启用/禁用、执行、日志)
- [x] 实现任务表单弹窗 (新增/编辑)
- [x] 实现表单校验规则
- [x] 实现系统任务的特殊处理 (指令只读、不可删除)

### Task 11: 实现执行日志页
- [x] 创建 `apps/lina-vben/apps/web-antd/src/views/monitor/job/log.vue`
- [x] 实现搜索区域 (任务名称、状态、时间范围)
- [x] 实现 VXE-Grid 表格
- [x] 实现详情抽屉
- [x] 实现从任务列表跳转并筛选功能

### Task 12: 配置路由
- [x] 修改 `apps/lina-vben/apps/web-antd/src/router/routes/modules/monitor.ts`
- [x] 添加定时任务路由
- [x] 添加执行日志路由

## 测试

### Task 13: 编写 E2E 测试用例
- [x] 创建 `hack/tests/e2e/monitor/TC0101-job-list.ts` (任务列表查询)
- [x] 创建 `hack/tests/e2e/monitor/TC0102-job-create.ts` (创建任务)
- [x] 创建 `hack/tests/e2e/monitor/TC0103-job-update.ts` (更新任务)
- [x] 创建 `hack/tests/e2e/monitor/TC0104-job-delete.ts` (删除任务)
- [x] 创建 `hack/tests/e2e/monitor/TC0105-job-status.ts` (启用/禁用任务)
- [x] 创建 `hack/tests/e2e/monitor/TC0106-job-run.ts` (手动执行任务)
- [x] 创建 `hack/tests/e2e/monitor/TC0107-job-log.ts` (执行日志查询)
- [x] 创建 `hack/tests/e2e/monitor/TC0108-job-system-protect.ts` (系统任务保护)
- [ ] 运行所有测试用例并确保通过

## 文档更新

### Task 14: 更新项目文档
- [ ] 更新 `README.md` 添加定时任务功能说明
- [ ] 更新 `CLAUDE.md` 添加定时任务相关规范 (如需要)

## 验收标准

- [x] 所有数据库表创建成功
- [x] 系统任务自动初始化
- [x] 任务列表页功能完整 (增删改查、启用/禁用、执行)
- [x] 执行日志页功能完整 (查询、详情)
- [x] 系统任务不可删除，指令不可修改
- [x] 单例执行模式正常工作 (基于分布式锁)
- [x] 执行次数限制正常工作
- [ ] 所有 E2E 测试用例通过
- [x] 前后端接口联调成功

## Feedback

- [x] **FB-1**: 创建任务表单组件 JobForm,修复页面加载报错
- [x] **FB-2**: 将定时任务菜单从"系统监控"移动到"系统管理"菜单下
- [x] **FB-3**: 修复组件导入方式和语法错误,确保页面能正常加载
