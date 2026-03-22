## 1. 数据库变更与基础设施

- [x] 1.1 创建 `manifest/sql/v0.7.0.sql`：新增 `sys_online_session`（MEMORY 引擎）和 `sys_server_monitor` 表的 DDL，新增系统监控菜单数据（"系统监控"一级菜单 + "在线用户"、"服务监控"子菜单）
- [x] 1.2 执行 `make init` 更新数据库，执行 `make dao` 生成 DAO/DO/Entity 文件

## 2. 在线用户后端实现

- [x] 2.1 定义会话存储抽象接口 `SessionStore`，实现基于 MySQL 的 `DBSessionStore`（创建、查询、删除、列表过滤）
- [x] 2.2 改造认证服务（`internal/service/auth/`）：登录成功后创建会话记录（包含 token_id、user_id、username、dept_name、ip、browser、os、login_time），登出时删除会话记录
- [x] 2.3 改造认证中间件（`internal/service/middleware/`）：在 JWT 校验后增加会话存在性检查，会话不存在时返回 401
- [x] 2.4 创建在线用户 API 定义（`api/monitor/v1/`）：`GET /monitor/online/list`（列表查询，支持 username/ip 过滤）、`DELETE /monitor/online/{tokenId}`（强制下线）
- [x] 2.5 实现在线用户 Controller 和 Service：列表查询、强制下线逻辑，执行 `make ctrl` 生成控制器骨架
- [x] 2.6 注册在线用户路由到 `internal/cmd/cmd.go`

## 3. 服务监控后端实现

- [x] 3.1 引入 `github.com/shirou/gopsutil/v4` 依赖，实现指标采集服务（CPU、内存、磁盘、网络、Go 运行时、服务器基本信息）
- [x] 3.2 实现定时采集任务：服务启动时立即采集一次，此后每 30 秒采集一次写入 `sys_server_monitor` 表，同时清理超过 1 小时的旧数据
- [x] 3.3 创建服务监控 API 定义（`api/monitor/v1/`）：`GET /monitor/server`（查询节点监控数据，支持 nodeName 过滤）
- [x] 3.4 实现服务监控 Controller 和 Service：读取数据库中各节点最新监控数据，执行 `make ctrl` 生成控制器骨架
- [x] 3.5 注册服务监控路由和定时采集任务到 `internal/cmd/cmd.go`

## 4. 在线用户前端实现

- [x] 4.1 创建前端 API 文件 `src/api/monitor/online/`：定义 `onlineList`、`forceLogout` 接口调用
- [x] 4.2 创建在线用户页面 `src/views/monitor/online/index.vue` 和 `data.ts`：搜索表单（用户名、IP）、VXE-Grid 表格（登录账号、部门、IP、登录地点、浏览器图标、OS 图标、登录时间、操作列）、工具栏在线人数统计、强制下线 Popconfirm
- [x] 4.3 新增系统监控路由模块 `src/router/routes/modules/monitor.ts`：配置"系统监控"一级路由及"在线用户"、"服务监控"子路由

## 5. 服务监控前端实现

- [x] 5.1 创建前端 API 文件 `src/api/monitor/server/`：定义 `getServerInfo` 接口调用
- [x] 5.2 创建服务监控页面 `src/views/monitor/server/index.vue` 及子组件：服务器信息卡片（Descriptions）、CPU 指标卡片（进度条）、内存指标卡片（进度条）、Go 运行时信息卡片、磁盘使用表格（含进度条）、网络流量信息展示
- [x] 5.3 实现多节点切换逻辑：多节点时显示节点选择下拉框，单节点时隐藏

## 6. E2E 测试

- [x] 6.1 创建 `hack/tests/e2e/monitor/TC0049-online-user-list.ts`：验证在线用户列表展示（表格列、在线人数统计）
- [x] 6.2 创建 `hack/tests/e2e/monitor/TC0050-online-user-search.ts`：验证按用户名、IP 搜索过滤功能
- [x] 6.3 创建 `hack/tests/e2e/monitor/TC0051-online-user-force-logout.ts`：验证强制下线交互（确认弹窗、下线后列表刷新）
- [x] 6.4 创建 `hack/tests/e2e/monitor/TC0052-server-monitor-page.ts`：验证服务监控页面展示（各指标卡片、磁盘表格、数据非空）
- [x] 6.5 运行全部 E2E 测试确认无回归

## Feedback

- [x] **FB-1**：`sys_online_session` 表新增 `last_active_time` 字段，登录时设置为当前时间，重新生成 DAO/DO/Entity
- [x] **FB-2**：Session Store 新增 `TouchOrValidate` 方法，通过 UPDATE 操作更新 `last_active_time` 并返回受影响行数判断会话是否存在；Auth 中间件改用此方法替代原有的 `Get` 查询
- [x] **FB-3**：Session Store 新增 `CleanupInactive` 方法，删除 `last_active_time` 超过阈值的会话记录；在 `cmd_http.go` 中启动定时清理任务（默认每5分钟执行一次）
- [x] **FB-4**：`config.yaml` 新增 `session.timeoutHour`（超时阈值，默认24）和 `session.cleanupMinute`（清理频率，默认5）配置项
