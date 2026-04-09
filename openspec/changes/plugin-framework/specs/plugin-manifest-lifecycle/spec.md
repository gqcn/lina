## ADDED Requirements

### Requirement: 插件目录与清单契约统一
系统 SHALL 为所有插件提供统一的目录结构与清单契约。源码插件 MUST 放置在 `apps/lina-plugins/<plugin-id>/` 目录下；当前运行时 `wasm` 插件 MUST 能从 `plugin.runtime.storagePath` 中被发现，并解析出与源码插件等价的 manifest 信息。

#### Scenario: 发现源码插件目录
- **WHEN** 宿主扫描 `apps/lina-plugins/` 下的插件目录
- **THEN** 仅将包含合法清单文件的目录识别为插件
- **AND** 每个插件的 `plugin-id` 在宿主范围内唯一
- **AND** 清单仅需包含插件基础信息与一级插件类型

#### Scenario: `plugin.yaml` 保持最小化基础元数据
- **WHEN** 宿主解析 `plugin.yaml`
- **THEN** 清单只要求 `id`、`name`、`version`、`type` 等基础字段
- **AND** 宿主不再要求 `schemaVersion`、`compatibility`、`entry` 等扩展元数据
- **AND** 菜单、权限、前端页面、`Slot` 和 SQL 文件位置优先按照目录与代码约定推导，而不是在清单中重复配置

#### Scenario: 清单一级类型只保留源码与运行时两类
- **WHEN** 宿主解析 `plugin.yaml` 中的 `type`
- **THEN** `type` 仅允许 `source` 或 `runtime`
- **AND** 当前仅 `wasm` 作为运行时插件的产物语义，不再作为一级插件类型
- **AND** 对历史上的 `wasm` 一级类型值，宿主在治理视角下统一按 `runtime` 处理

#### Scenario: 安装运行时插件产物
- **WHEN** 管理员上传一个 `wasm` 文件安装运行时插件
- **THEN** 宿主能够解析出与源码模式一致的插件标识、名称、版本与一级插件类型
- **AND** 对缺少这些基础字段的运行时插件拒绝安装
- **AND** 宿主将上传产物写入 `plugin.runtime.storagePath/<plugin-id>.wasm`

#### Scenario: 运行时产物使用独立存储目录
- **WHEN** 宿主发现、上传或同步一个运行时 `wasm` 插件产物
- **THEN** 运行时产物 MUST 使用 `plugin.runtime.storagePath/<plugin-id>.wasm` 作为宿主侧规范落盘路径
- **AND** 宿主不得再依赖 `apps/lina-plugins/<plugin-id>/plugin.yaml` 作为 runtime 发现入口
- **AND** 运行时样例插件的可读源码目录 SHOULD 与源码插件一样继续收敛在 `backend/`、`frontend/` 与 `manifest/` 下维护

### Requirement: 插件生命周期状态机可治理
系统 SHALL 为插件提供可审计的生命周期状态机，并按源码插件与运行时插件区分生命周期语义。

#### Scenario: 源码插件随宿主编译集成
- **WHEN** 宿主编译源码插件所在的源码树并生成 Lina 二进制
- **THEN** 源码插件的后端 Go 代码与宿主源码一起完成编译
- **AND** 源码插件在插件注册表中视为已集成，不需要额外安装步骤
- **AND** 管理员只需要管理源码插件的启用与禁用状态

#### Scenario: 源码插件首次同步后默认启用
- **WHEN** 宿主首次发现一个源码插件并将其写入插件注册表
- **THEN** 该源码插件默认处于“已集成且已启用”状态
- **AND** 宿主后续同步不会覆盖管理员对该源码插件做出的显式禁用操作

#### Scenario: 安装运行时插件
- **WHEN** 管理员安装一个合法的 `wasm` 运行时插件
- **THEN** 宿主创建插件安装记录与当前版本记录
- **AND** 宿主按清单依次处理迁移、资源注册、权限接入与前后端装载准备
- **AND** 插件在显式启用前不会对普通用户可见

#### Scenario: 禁用插件
- **WHEN** 管理员将已启用插件切换为禁用状态
- **THEN** 宿主停止该插件的 Hook、Slot、页面与菜单暴露
- **AND** 宿主保留插件业务数据、角色授权关系与安装记录
- **AND** 插件重新启用后可以恢复既有治理关系

#### Scenario: 卸载运行时插件
- **WHEN** 管理员卸载一个运行时插件
- **THEN** 宿主移除该插件在宿主侧注册的菜单、资源引用、运行时产物与挂载信息
- **AND** 宿主默认不删除插件自己的业务数据表或业务数据
- **AND** 宿主保留卸载审计信息

#### Scenario: 升级插件
- **WHEN** 管理员为已安装插件安装更高版本的 release
- **THEN** 宿主为插件创建新的 release 记录与代际信息
- **AND** 旧 release 在新 release 生效前保持可回退
- **AND** 升级失败时宿主能够回滚到上一个可用 release

#### Scenario: 源码插件不暴露安装卸载操作
- **WHEN** 管理员查看源码插件的插件管理操作项
- **THEN** 宿主不会为源码插件展示安装或卸载操作
- **AND** 源码插件仅暴露同步发现、启用和禁用等适用操作

### Requirement: 插件资源归属与迁移记录可追踪
系统 SHALL 记录插件对宿主资源与迁移的占用关系，以支持卸载、重装、升级、审计与故障恢复。

#### Scenario: 插件注册宿主资源
- **WHEN** 插件在安装期间创建或声明菜单、权限、配置、字典、静态资源或其他宿主治理资源
- **THEN** 宿主记录该资源与插件、release 的归属关系
- **AND** 这些引用关系可以被查询、审计和用于卸载清理

#### Scenario: 执行插件迁移
- **WHEN** 插件安装或升级需要执行 SQL 或其他迁移步骤
- **THEN** 宿主记录每个迁移项的执行顺序、版本、校验摘要、执行结果与时间
- **AND** 同一个 release 的同一个迁移项不会被重复执行

#### Scenario: 插件版本 SQL 命名与目录约束
- **WHEN** 插件在 `manifest/sql/` 目录下提供安装阶段 SQL
- **THEN** 安装 SQL 文件 MUST 使用与宿主一致的命名格式 `{序号}-{当前迭代名称}.sql`
- **AND** 这些安装 SQL 文件 MUST 放在插件的 `manifest/sql/` 根目录下，供宿主按顺序扫描执行
- **AND** 插件卸载 SQL MUST 独立放在 `manifest/sql/uninstall/` 目录下
- **AND** 宿主初始化顺序执行流程 MUST 只扫描 `manifest/sql/` 根目录，不得误执行 `manifest/sql/uninstall/` 下的卸载 SQL

#### Scenario: 插件菜单安装不依赖整型菜单 ID
- **WHEN** 插件通过安装 SQL 写入宿主菜单与按钮权限
- **THEN** 菜单记录 MUST 使用 `menu_key` 作为菜单稳定标识
- **AND** 父子关系 MUST 通过 `menu_key` 解析真实 `parent_id`，而不是写死固定整型 `parent_id`
- **AND** 插件安装、升级与卸载流程 MUST 不依赖固定整型 `id`

#### Scenario: 安装过程部分失败
- **WHEN** 插件在迁移、资源注册或产物准备过程中任一步骤失败
- **THEN** 宿主将插件状态标记为失败或待人工介入
- **AND** 宿主回滚尚未生效的宿主治理资源
- **AND** 宿主保留失败上下文供后续诊断
