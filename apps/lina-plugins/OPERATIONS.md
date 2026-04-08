# Lina 插件运维与 Review 指南

本文档只描述**当前仓库已经落地的真实能力**，用于帮助维护者和人工 reviewer 在执行插件同步、启用、禁用、安装、卸载相关操作后，快速确认宿主侧元数据是否保持一致。

## 适用范围

当前仓库仍然以**第一期源码插件底座**为主，`apps/lina-plugins/` 下也仅保留 `plugin-demo` 作为唯一插件样例。运行时插件完整能力尚未交付，但宿主已经先补齐了后续阶段会复用的元数据表：

- `sys_plugin`
- `sys_plugin_release`
- `sys_plugin_migration`
- `sys_plugin_resource_ref`
- `sys_plugin_node_state`

这意味着，即使二三期 `runtime wasm`、多节点热更新等能力还没有完成，reviewer 也已经可以通过数据库直接看到：

- 宿主当前识别到的插件注册状态
- 宿主认为当前生效的是哪个插件版本
- 插件安装/卸载迁移的执行记录与抽象执行键
- 宿主从插件目录中发现了哪些资源类型，以及对应的数量摘要与抽象标识
- 当前节点对插件状态的本地观测结果

## 当前生命周期模型

## 源码插件

源码插件通过 `apps/lina-plugins/<plugin-id>/plugin.yaml` 被宿主发现。

当前实际行为如下：

1. 宿主扫描插件目录。
2. 宿主校验 `plugin.yaml`、`go.mod`、`backend/plugin.go`、SQL 命名规则以及前端页面/`Slot` 目录约定。
3. 宿主同步 `sys_plugin` 插件注册表。
4. 宿主将当前清单基础字段和资源数量摘要写入 `sys_plugin_release`。
5. 宿主将目录发现到的资源类型与抽象 owner 信息写入 `sys_plugin_resource_ref`。
6. 宿主将当前节点的插件状态投影写入 `sys_plugin_node_state`。

当前约束如下：

- 源码插件视为“随宿主编译即已集成”，不走安装/卸载流程。
- 源码插件禁用后，只会隐藏路由、菜单、页面和 `Slot`，不会删除历史业务数据。
- 当前仓库只保留 `plugin-demo` 一个插件样例，不再新增其他插件目录作为模板或示例。

## 运行时插件

运行时插件的完整执行模型仍在后续阶段，但宿主服务层已经先补齐了一部分基础记录能力。

当前已经落地的行为：

1. 安装/卸载生命周期会把 SQL 执行结果记录到 `sys_plugin_migration`。
2. 安装完成后会同步当前版本快照、资源引用和节点状态。
3. 卸载完成后会清理当前版本的资源引用，并刷新节点状态。

当前**尚未**落地的行为：

- 上传并校验真实 `wasm` 产物
- 运行时前端静态资源托管
- 真正的运行时装载、热升级、回滚和多节点代际切换

## Review 检查清单

建议人工 review 按下面顺序核对。

## 1. 插件注册表

核对 `sys_plugin`：

- `plugin_id`、`name`、`version`、`type` 是否与 `plugin.yaml` 一致
- `installed`、`status` 是否与本次操作预期一致
- `manifest_path` 是否指向正确的插件清单位置

建议查询：

```sql
SELECT plugin_id, version, type, installed, status, manifest_path, installed_at, enabled_at, disabled_at
FROM sys_plugin
ORDER BY plugin_id;
```

## 2. 发布快照

核对 `sys_plugin_release`：

- `release_version` 是否与当前生效版本一致
- `status` 是否符合当前生命周期动作（如 `active`、`installed`、`uninstalled`）
- `manifest_snapshot` 中应包含基础清单字段和资源数量摘要，而不是具体文件路径
- `manifest_snapshot` 中的 `installSqlCount`、`uninstallSqlCount`、`frontendPageCount`、`frontendSlotCount` 是否与目录实际情况一致

建议查询：

```sql
SELECT plugin_id, release_version, type, status, manifest_path, package_path, checksum
FROM sys_plugin_release
ORDER BY plugin_id, release_version;
```

## 3. 迁移执行记录

在运行时插件安装/卸载链路中，核对 `sys_plugin_migration`：

- `phase` 只能是 `install` 或 `uninstall`
- `migration_key` 应为抽象执行键，例如 `install-step-001`
- `status = succeeded` 表示执行成功
- 如果 SQL 内容变更，`checksum` 也应随之变化

建议查询：

```sql
SELECT plugin_id, release_id, phase, migration_key, execution_order, status, error_message, executed_at
FROM sys_plugin_migration
ORDER BY plugin_id, release_id, phase, execution_order;
```

## 4. 资源引用

核对 `sys_plugin_resource_ref`：

- 每个发现到的资源都应有稳定的 `resource_type` 与 `resource_key`
- `resource_path` 当前应保持为空字符串，或仅作为宿主抽象定位补充信息，不允许保存具体前端/SQL 文件路径
- `remark` 应描述数量摘要或宿主识别结论，便于 review 快速判断发现结果
- 当前源码插件快照至少应覆盖：
  - `manifest`
  - `backend_entry`
  - `install_sql`
  - `uninstall_sql`
  - `frontend_page`
  - `frontend_slot`

建议查询：

```sql
SELECT plugin_id, release_id, resource_type, resource_key, resource_path, owner_type, owner_key, remark
FROM sys_plugin_resource_ref
ORDER BY plugin_id, release_id, resource_type, resource_key;
```

## 5. 节点状态投影

核对 `sys_plugin_node_state`：

- `node_key` 应为当前宿主节点主机名
- `release_id` 应指向当前节点观测到的发布记录
- `desired_state` 与 `current_state` 应与 `installed/enabled` 组合保持一致
- `generation` 不应回退；当前实现至少与 `release_id` 对齐或更大

当前映射规则：

- `installed = 0` -> `uninstalled`
- `installed = 1` 且 `enabled = 0` -> `installed`
- `installed = 1` 且 `enabled = 1` -> `enabled`

建议查询：

```sql
SELECT plugin_id, release_id, node_key, desired_state, current_state, generation, last_heartbeat_at, error_message
FROM sys_plugin_node_state
ORDER BY plugin_id, node_key;
```

## 常见 Review 问题

## 为什么资源引用不直接来自 `plugin.yaml`？

因为当前插件设计明确要求 `plugin.yaml` 保持最小化。页面、`Slot`、SQL 等信息应由真实目录结构推导，而不是在清单里重复维护一份配置，避免双真相源。

## 为什么宿主还要额外保存 `manifest_snapshot`？

因为人工 review 不能只依赖磁盘上的当前文件，还需要一个“宿主在某次同步时到底看到了什么”的持久化快照。但当前快照只保存基础清单字段和数量摘要，不保存具体 SQL 文件路径或前端源码路径，避免把框架实现细节硬编码进治理表。

## 为什么多节点运行时能力还没做完，就先加了 `sys_plugin_node_state`？

因为后续多节点阶段一定需要宿主持久化节点侧观测结果。现在先把表结构补齐，可以避免后续继续改底层元数据模型，也能让当前人工 review 立即获得一个节点视角的检查入口。

## 当前已知限制

- 宿主当前还没有把“源码插件目录被删除”自动回收进注册表同步逻辑中。
- 宿主当前还没有把插件菜单 SQL 进一步解析成 `menu_key` 级别的宿主资源引用记录。
- 运行时插件的真实产物上传、装载隔离和热升级仍未实现。
- 仓库当前不提供插件打包脚本；这是有意保留的约束，避免出现与真实实现脱节的辅助脚本。
