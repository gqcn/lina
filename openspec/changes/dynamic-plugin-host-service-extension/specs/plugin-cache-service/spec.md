## ADDED Requirements

### Requirement: 动态插件通过命名缓存空间访问宿主缓存

系统 SHALL 为动态插件提供受治理的缓存服务，插件只能通过宿主授权的命名缓存空间访问缓存，而不能直接获取 Redis 或其他缓存客户端。

#### Scenario: 插件访问授权缓存空间

- **WHEN** 插件调用缓存服务执行`get`、`set`、`delete`、`incr`或`expire`
- **THEN** 宿主仅允许访问当前插件已授权的`host-cache`资源
- **AND** 宿主按该缓存空间的命名规则和 TTL 策略执行操作

#### Scenario: 插件尝试访问未授权缓存空间

- **WHEN** 插件调用一个未授权的缓存空间
- **THEN** 宿主拒绝该调用
- **AND** 宿主不向 guest 暴露底层缓存连接信息
