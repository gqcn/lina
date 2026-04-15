## Why

`lina-core` 当前的时长类配置同时存在两种表达方式：一类使用整数并把单位写进字段名（如 `expireHour`、`cleanupMinute`），另一类直接使用带单位的 duration 字符串（如 `30s`）。这种混用方式让配置含义不统一，也迫使业务代码在多个位置重复做小时、分钟、秒的换算，增加后续维护和扩展成本。

## What Changes

- 将 `jwt`、`session`、`monitor` 下的时长配置统一为带单位的 duration 字符串。
- 将整数+单位后缀的配置键调整为统一语义命名：`jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval`。
- 配置服务统一将这些配置解析为 `time.Duration`，业务层直接消费解析结果，不再自行做单位换算。
- 旧整数配置键不再保留任何兼容逻辑，配置文件、代码和文档仅保留新的 duration 写法。
- 更新相关测试，覆盖新配置、默认值和非法 duration 输入行为。

## Capabilities

### New Capabilities
- 无

### Modified Capabilities
- `user-auth`: JWT Token 有效期配置改为使用 duration 字符串键 `jwt.expire`。
- `online-user`: 在线会话超时阈值与清理周期改为使用 duration 字符串键 `session.timeout`、`session.cleanupInterval`。
- `server-monitor`: 服务监控采集周期改为使用 duration 字符串键 `monitor.interval`，并保持保留倍数配置不变。

## Impact

- 影响后端配置文件：`apps/lina-core/manifest/config/config.yaml`、`apps/lina-core/manifest/config/config.template.yaml`
- 影响配置读取与消费代码：`apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/auth/`、`apps/lina-core/internal/service/role/`、`apps/lina-core/internal/service/cron/`
- 影响相关单元测试与 OpenSpec 规范文档
