## 1. 配置与解析收敛

- [x] 1.1 更新 `lina-core` 默认配置与模板配置，使用新的 duration 字符串键 `jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval`
- [x] 1.2 重构 `internal/service/config` 的时长配置读取逻辑，统一返回 `time.Duration`

## 2. 业务消费改造

- [x] 2.1 调整认证与权限缓存逻辑，改为直接消费 JWT 与会话的 `time.Duration` 配置
- [x] 2.2 调整在线会话清理与服务监控采集/清理定时任务，统一基于 `time.Duration` 调度与计算阈值

## 3. 验证与文档

- [x] 3.1 补充配置相关单元测试，覆盖默认值、新配置解析与非法 duration 输入行为
- [x] 3.2 更新本次 OpenSpec 变更任务状态并运行相关测试，确认配置迁移无回归

## Feedback

- [x] **FB-1**: 移除旧配置键兼容逻辑，按全新项目方案仅保留新的 duration 配置实现
