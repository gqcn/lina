## Context

`lina-core` 的时长配置当前分散在多个模块中读取和换算：`jwt.expireHour`、`session.timeoutHour`、`session.cleanupMinute`、`monitor.intervalSeconds` 使用整数表达，`cluster.election.lease`、`cluster.election.renewInterval` 则直接使用 duration 字符串。前者要求业务层自行拼接 `time.Hour`、`time.Minute`、`time.Second`，导致相同类型的配置在配置层、服务层、定时任务层呈现出不同的处理方式。

这次变更横跨配置定义、配置读取、认证、在线会话和服务监控三个模块，因此需要先明确统一的解析策略和调度方式。

## Goals / Non-Goals

**Goals:**
- 统一 `jwt`、`session`、`monitor` 的时长配置写法，全部使用带单位的 duration 字符串。
- 在配置服务层把相关配置统一解析为 `time.Duration`，业务层直接消费。
- 更新默认配置文件、测试和规范文档，使新写法成为唯一推荐写法。

**Non-Goals:**
- 不修改 `cluster.election.lease`、`cluster.election.renewInterval` 的现有语义与键名。
- 不改动数据库结构、外部 API 路径或响应格式。
- 不将所有非时长数值配置统一改造成字符串，诸如 `monitor.retentionMultiplier` 这类倍数配置保持整数。

## Decisions

### 决策 1：配置文件统一使用 duration 字符串，代码内部统一使用 `time.Duration`
- 方案：新增并推荐 `jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval`，值使用 `24h`、`5m`、`30s` 这类格式；配置服务返回 `time.Duration`。
- 原因：配置值自身携带单位，避免字段名重复承载单位语义；业务层直接消费 `time.Duration`，可以消除重复换算代码。
- 备选方案：继续保留整数配置并仅重命名字段。否决原因是业务层仍然需要换算单位，混用问题没有真正消除。

### 决策 2：不保留任何旧键兼容逻辑
- 方案：配置服务只解析 `jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval` 这些新键，旧键不再参与读取。
- 原因：当前项目属于全新项目，没有历史配置负担，保留兼容逻辑只会增加复杂度并模糊最终配置标准。
- 备选方案：兼容读取旧键。否决原因是会把过渡性复杂度永久留在实现中，不符合“全新项目、无历史债务”的约束。

### 决策 3：配置层显式解析 duration，而不是依赖通用 `Scan`
- 方案：为时长配置增加显式解析逻辑，统一解析 duration 字符串并做必要校验。
- 原因：显式解析更直观，也方便在配置层集中校验最小粒度与非法输入。
- 备选方案：继续用结构体 `Scan`。否决原因是错误信息和约束控制不够集中。

### 决策 4：定时任务改用 duration 风格调度表达
- 方案：会话清理与监控采集调度统一基于解析后的 duration 生成调度方式，优先使用 `@every <duration>` 形式，保留现有主节点判定逻辑不变。
- 原因：和新的 duration 配置天然匹配，避免把 `time.Duration` 再拆回分钟或秒去拼六段 cron 表达式。
- 备选方案：继续拼 cron 表达式。否决原因是需要再次手工提取分钟或秒，且不利于支持未来更灵活的间隔值。

## Risks / Trade-offs

- [风险] duration 字符串配置写错单位或漏写单位会导致启动期解析失败 → 缓解：提供明确默认值与错误信息，并在测试中覆盖非法配置输入。
- [风险] `@every` 调度与原 cron 表达式存在首轮触发时机差异 → 缓解：保留监控启动后立即采集的现有逻辑，并让周期任务只负责后续调度。
- [取舍] 不提供旧键兼容会让错误配置在启动期尽早暴露 → 换取实现更简单、配置标准更统一。

## Migration Plan

1. 将配置模板与默认配置切换为新的 duration 字符串键。
2. 在配置服务中实现新的 duration 解析与校验逻辑。
3. 修改认证、在线会话和监控模块，使其只消费 `time.Duration`。
4. 增补单元测试，覆盖默认值、新键解析和非法 duration 输入行为。
5. 如需回滚，可恢复上一版本代码。

## Open Questions

- 暂无。当前方案已明确采用纯新项目配置路径。
