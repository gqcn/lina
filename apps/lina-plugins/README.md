# Lina 插件机制设计与开发指南

`apps/lina-plugins/`既是当前插件机制的一期设计文档，也是后续插件开发时的统一参考入口。本文档只描述**仓库中已经落地的真实实现**，同时明确哪些内容仍属于后续规划，避免设计文档、代码实现和开发习惯之间出现偏差。

## 文档定位

本文档同时服务两类读者：

- 插件开发者：需要按照当前约定创建、接入、调试和维护插件。
- 人工 reviewer：需要快速确认某个插件是否符合当前插件框架的设计边界和接入规范。

因此本文档的目标不是“讲概念”，而是回答以下问题：

- 当前插件机制到底支持什么，不支持什么。
- 一个新的源码插件应该放在哪里，哪些文件是必需的。
- 宿主如何发现插件页面、`Slot`、SQL 和后端注册入口。
- `plugin.yaml` 为什么保持最小化，以及哪些字段明确不允许再放进去。
- 开发者在提交插件前应该自查哪些关键点。

## 当前范围

当前仓库已经落地的是**第一期：源码插件底座**。插件机制的能力边界如下：

| 能力 | 当前状态 | 说明 |
|------|------|------|
| `source`源码插件 | 已实现 | 插件目录位于`apps/lina-plugins/<plugin-id>/`，随宿主一起编译、打包和交付 |
| `runtime`运行时插件 | 规划中 | 当前数据库、接口和规格中保留了治理入口，但完整安装链路尚未交付 |
| 插件管理页 | 已实现 | 支持源码插件同步、启用、禁用与治理联动 |
| 后端扩展点 | 已实现 | 通过`pluginhost`发布的回调式扩展点接入 |
| 前端页面接入 | 已实现 | 扫描`frontend/pages/**/*.vue`并挂到宿主运行时页 |
| 前端`Slot`接入 | 已实现 | 扫描`frontend/slots/**/*.vue`并挂到宿主公开插槽 |
| 插件安装 SQL | 已实现 | 通过`manifest/sql/*.sql`目录约定发现 |
| 插件卸载 SQL | 已实现 | 通过`manifest/sql/uninstall/*.sql`目录约定发现 |
| 脚手架脚本 | 未提供 | 当前不再提供`hack/plugin`下的脚本，避免生成物与真实实现脱节 |

补充说明：

- 运维与 review 说明已整理到 [OPERATIONS.md](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/OPERATIONS.md)。
- 当前仓库继续只保留 `plugin-demo` 作为唯一插件样例，不再新增额外插件模板目录；新插件目录规范以本文档和 `plugin-demo` 为准。

## 设计原则

当前插件机制遵循以下原则：

### 约定优于配置

- 前端页面位置、前端`Slot`位置、安装 SQL 和卸载 SQL 都通过固定目录约定发现。
- `plugin.yaml` 不再重复声明这些信息，避免同一份事实在多个位置维护。

### 单一真相源

- 菜单、权限和父子关系以 SQL 为单一真相源。
- 后端扩展能力以插件代码注册为单一真相源。
- 前端页面和`Slot`以真实源码文件为单一真相源。

### 显式接线

- 当前源码插件的后端接线方式不是脚本生成，也不是隐式自动装配。
- 开发者需要显式维护`apps/lina-plugins/lina-plugins.go`，让宿主编译期导入插件后端包。
- 这样做的目的是让接线关系清晰、可 grep、可 review、可追踪。

### 设计与实现一致

- 文档中不再保留已经移除的元数据模型和自动化脚本描述。
- 对未来能力的描述会明确标注为“规划中”，不能和“已实现”混写。

## 插件类型

当前插件一级类型只保留两类：

| 类型 | 含义 | 当前状态 |
|------|------|------|
| `source` | 源码插件，目录在`apps/lina-plugins/<plugin-id>/` | 已实现 |
| `runtime` | 运行时插件，面向后续热安装与热升级 | 规划中 |

重要说明：

- 当前虽然规格中保留了`runtime`类型，但仓库真正已经闭环验证的是`source`源码插件。
- 历史上把`wasm`当一级类型的设计已经收敛掉了。当前治理视角只区分`source`和`runtime`。
- 如果当前要开发新插件，默认应按照`source`源码插件方式开发。

## 源码插件生命周期

源码插件和运行时插件的生命周期语义并不相同。源码插件当前遵循下表：

| 动作 | 源码插件行为 |
|------|------|
| 发现 | 宿主扫描`apps/lina-plugins/*/plugin.yaml`识别插件 |
| 同步 | 宿主同步`sys_plugin`记录，保持插件列表和实际目录一致 |
| 安装 | 不提供。源码插件视为随宿主编译即已集成 |
| 卸载 | 不提供。移除源码插件需要修改源码目录和注册关系后重新构建 |
| 启用 | 已支持。启用后路由、菜单、页面和`Slot`恢复生效 |
| 禁用 | 已支持。禁用后路由、菜单、页面和`Slot`隐藏或拒绝访问 |

这意味着：

- 插件管理页中，源码插件不应出现“安装”“卸载”按钮。
- 新增一个源码插件后，如果目录、清单和注册关系都正确，宿主同步后会把它视为已集成插件。
- 禁用插件不会删除已有业务数据；重新启用后，应能恢复原有治理关系。

## 目录结构

当前源码插件统一放在`apps/lina-plugins/<plugin-id>/`下。推荐目录如下：

```text
apps/lina-plugins/
  README.md
  lina-plugins.go
  <plugin-id>/
    go.mod
    plugin.yaml
    README.md
    backend/
      plugin.go
      api/
      internal/
        controller/
      service/
    frontend/
      pages/
        *.vue
      slots/
        <slot-key>/
          *.vue
    manifest/
      sql/
        001-<iteration-name>.sql
        uninstall/
          001-<iteration-name>.sql
```

各目录职责如下：

| 路径 | 作用 | 是否必需 |
|------|------|------|
| `apps/lina-plugins/lina-plugins.go` | 宿主源码插件后端导入注册表 | 是 |
| `<plugin-id>/go.mod` | 插件独立 Go 模块声明 | `source`插件必需 |
| `<plugin-id>/plugin.yaml` | 插件最小元数据清单 | 是 |
| `<plugin-id>/README.md` | 插件自身说明文档 | 强烈建议 |
| `<plugin-id>/backend/plugin.go` | 插件后端注册入口 | `source`插件必需 |
| `<plugin-id>/backend/api/` | 插件 API 定义 | 按需 |
| `<plugin-id>/backend/internal/controller/` | 插件控制器实现 | 按需 |
| `<plugin-id>/backend/service/` | 插件服务层实现 | 按需 |
| `<plugin-id>/frontend/pages/` | 插件页面源码目录 | 有页面时必需 |
| `<plugin-id>/frontend/slots/` | 插件`Slot`源码目录 | 有`Slot`时必需 |
| `<plugin-id>/manifest/sql/` | 插件安装 SQL 目录 | 有安装 SQL 时必需 |
| `<plugin-id>/manifest/sql/uninstall/` | 插件卸载 SQL 目录 | 有卸载 SQL 时必需 |

## 元数据底座

为了让后续人工 review 不必只依赖日志，宿主当前会把插件治理元数据同步到以下表中：

| 表名 | 当前用途 |
|------|------|
| `sys_plugin` | 插件注册表，记录插件基础状态 |
| `sys_plugin_release` | 记录插件版本、清单基础信息和资源数量摘要快照 |
| `sys_plugin_migration` | 记录安装/卸载迁移的执行结果与抽象执行键 |
| `sys_plugin_resource_ref` | 记录宿主发现到的抽象资源类型、稳定标识与摘要说明 |
| `sys_plugin_node_state` | 记录当前节点对插件状态的观测结果 |

这些表的目标不是把二三期能力一次性做完，而是先把后续 runtime 生命周期需要的宿主元数据底座稳定下来。

同时需要明确当前持久化边界：

- 宿主会按目录约定扫描 SQL、页面和 `Slot`，但这些具体文件路径只用于校验与执行，不写入插件治理表。
- `manifest_snapshot` 只保存基础清单字段、是否声明清单以及各类资源数量摘要。
- `sys_plugin_resource_ref` 只保存抽象资源键、owner 标识和 summary remark，不保存具体前端文件路径或 SQL 文件路径。
- `sys_plugin_migration` 只保存类似 `install-step-001` 的抽象迁移执行键，不保存具体 SQL 相对路径。

当前插件管理页已经基于这些表补齐了以下治理摘要字段，便于人工 review：

| 字段 | 说明 |
|------|------|
| `releaseVersion` | 宿主当前视角下的生效版本号 |
| `lifecycleState` | 生命周期状态键，如 `source_enabled`、`runtime_installed` |
| `nodeState` | 当前节点观测状态，如 `enabled`、`installed`、`uninstalled` |
| `resourceCount` | 当前生效版本登记的资源引用数量 |
| `migrationState` | 最近一次迁移结果，如 `none`、`succeeded`、`failed` |

## `plugin.yaml`

### 设计目标

当前`plugin.yaml`故意保持最小化。它的职责只有两类：

- 声明“这个目录是一个插件”。
- 提供插件在治理侧展示和校验所需的基础身份信息。

它**不再负责**：

- 声明页面入口。
- 声明前端`Slot`。
- 声明 SQL 文件列表。
- 声明菜单前缀、权限前缀或菜单结构。
- 声明宿主兼容矩阵、脚本入口、打包入口。

### 推荐示例

```yaml
id: plugin-demo
name: 示例插件
version: v0.1.0
type: source
description: 提供插件扫描、状态管理、左侧菜单页面、前端 Slot 与公开/受保护路由示例的源码插件
author: lina-team
homepage: https://example.com/lina/plugins/plugin-demo
license: Apache-2.0
```

### 字段说明

| 字段 | 是否必填 | 说明 |
|------|------|------|
| `id` | 是 | 插件稳定标识，必须使用`kebab-case`，且在宿主范围内唯一 |
| `name` | 是 | 插件显示名称 |
| `version` | 是 | 插件版本号，必须使用`semver`格式；本文档示例统一使用带`v`前缀的写法 |
| `type` | 是 | 当前仅允许`source`或`runtime` |
| `description` | 否 | 插件简要描述，建议明确功能边界 |
| `author` | 否 | 插件作者或团队标识 |
| `homepage` | 否 | 插件主页或项目地址 |
| `license` | 否 | 插件许可信息 |

### 宿主校验规则

宿主当前会对`plugin.yaml`做以下校验：

| 校验项 | 规则 |
|------|------|
| `id` 非空 | 缺失则判定清单非法 |
| `id` 格式 | 必须匹配`^[a-z0-9]+(?:-[a-z0-9]+)*$` |
| `id` 唯一性 | 不允许两个插件目录使用同一个`id` |
| `name` 非空 | 缺失则判定清单非法 |
| `version` 非空 | 缺失则判定清单非法 |
| `version` 格式 | 必须满足`semver`格式，例如`v0.1.0`；宿主当前同时兼容不带`v`前缀的写法 |
| `type` 合法性 | 仅允许`source`或`runtime` |
| `source`目录完整性 | `source`插件必须存在`go.mod`和`backend/plugin.go` |

### 明确不再允许的字段

以下字段已经被当前设计明确淘汰，不应再写入`plugin.yaml`：

- `schemaVersion`
- `compatibility`
- `entry`
- `capabilities`
- `resources`
- `metadata`

这些字段被移除的原因是它们会把以下信息重复建模：

- SQL 文件路径，本来就可以从固定目录推导。
- 前端页面和`Slot`文件，本来就可以从真实源码目录推导。
- 菜单和权限信息，本来就应该以 SQL 为真相源。
- 路由和扩展点接入，本来就应该以插件代码注册为真相源。

## 后端接入

### 总体模型

源码插件的后端接入是“插件目录内实现 + `pluginhost`注册 + 宿主显式导入”三段式模型：

1. 插件在`backend/plugin.go`里创建并注册`SourcePlugin`。
2. 插件通过`pluginhost`向宿主注册路由、Hook、过滤器等回调。
3. 宿主通过`apps/lina-plugins/lina-plugins.go`匿名导入插件后端包，让其`init()`逻辑参与宿主编译产物。

### 宿主导入注册表

当前导入注册表文件是：

```go
package linaplugins

import (
	_ "lina-plugin-demo/backend"
)
```

新增插件时，开发者需要手工追加匿名导入，例如：

```go
package linaplugins

import (
	_ "lina-plugin-demo/backend"
	_ "lina-plugin-foo/backend"
)
```

这是当前源码插件后端接入的**唯一显式接线点**。

### `backend/plugin.go` 最小示例

```go
package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
)

const pluginID = "plugin-demo"

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	middlewares := registrar.Middlewares()

	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)
	})
	return nil
}
```

### 当前已发布的后端扩展点

宿主当前已经正式发布的后端扩展点如下：

| Go 常量 | Canonical 值 | 类型 | 支持模式 |
|------|------|------|------|
| `ExtensionPointAuthLoginSucceeded` | `auth.login.succeeded` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointAuthLoginFailed` | `auth.login.failed` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointAuthLogoutSucceeded` | `auth.logout.succeeded` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginInstalled` | `plugin.installed` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginEnabled` | `plugin.enabled` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginDisabled` | `plugin.disabled` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginUninstalled` | `plugin.uninstalled` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointSystemStarted` | `system.started` | 事件 Hook | `blocking`、`async` |
| `ExtensionPointHTTPRouteRegister` | `http.route.register` | 注册点 | `blocking` |
| `ExtensionPointHTTPRequestAfterAuth` | `http.request.after-auth` | 注册点 | `blocking` |
| `ExtensionPointCronRegister` | `cron.register` | 注册点 | `blocking` |
| `ExtensionPointMenuFilter` | `menu.filter` | 注册点 | `blocking` |
| `ExtensionPointPermissionFilter` | `permission.filter` | 注册点 | `blocking` |

开发约束：

- 事件 Hook 可以使用`blocking`或`async`。
- 注册式扩展点当前只允许`blocking`。
- 如果为扩展点声明了不支持的执行模式，宿主会在注册阶段拒绝。

### `RouteRegistrar` 能力

插件路由注册当前通过`RouteRegistrar`完成。它提供两类能力：

| 能力 | 说明 |
|------|------|
| `Group(prefix, fn)` | 在宿主插件路由根分组下创建路由分组 |
| `Middlewares()` | 获取宿主已发布的中间件目录 |

当前可供插件组合的宿主中间件包括：

- `NeverDoneCtx()`
- `HandlerResponse()`
- `CORS()`
- `Ctx()`
- `Auth()`
- `OperLog()`

重要语义：

- 插件路由本身受插件启停状态保护。插件被禁用后，宿主会在路由入口处直接拒绝访问。
- 宿主不会为插件自动追加固定前缀。插件自己决定是否挂到`/api/v1`或其他前缀下。
- 同一个插件可以在一次注册中拆分多个分组，分别挂载匿名和鉴权路由。

### `CronRegistrar` 能力

如果插件需要注册定时任务，可以使用`RegisterCron`和`CronRegistrar`：

| 能力 | 说明 |
|------|------|
| `Add(ctx, pattern, name, handler)` | 注册一个受插件启停保护的定时任务 |
| `IsPrimaryNode()` | 返回当前节点是否为主节点 |

建议：

- 如果任务只应该在主节点执行，插件应自行在回调内通过`IsPrimaryNode()`做判断。
- 定时任务的业务逻辑应放在插件自己的服务层，不要把大段业务逻辑堆在注册回调里。

### 插件后端资源声明

当前源码插件仍支持通过`RegisterResource`声明后端资源，以便复用宿主通用资源查询接口：

- 资源声明在插件代码中完成，而不是在`plugin.yaml`中配置。
- 资源查询统一走宿主的`GET /plugins/{id}/resources/{resource}`契约。

如果插件不需要暴露这类统一资源接口，可以完全不注册。

## 前端页面接入

### 目录约定

插件页面统一放在：

```text
frontend/pages/**/*.vue
```

宿主构建时会扫描这些页面源码，并将其挂载到插件运行时页面容器中。

### `pluginPageMeta`

页面文件可以通过导出`pluginPageMeta`提供显式元数据，例如：

```vue
<script lang="ts">
export const pluginPageMeta = {
  routePath: 'plugin-demo-sidebar-entry',
  title: '插件示例',
};
</script>
```

当前支持的页面元数据字段如下：

| 字段 | 是否必填 | 说明 |
|------|------|------|
| `pluginId` | 否 | 不传时默认从文件路径推导 |
| `routePath` | 否 | 不传时宿主会根据文件路径自动推导 |
| `title` | 否 | 不传时默认使用`routePath` |

### 默认路由推导规则

如果页面没有显式声明`routePath`，宿主会根据文件路径推导。规则是：

- 取插件 ID。
- 取`frontend/pages/`后的相对路径。
- 将路径中的`/`替换为`-`。
- 将路径中的`_`替换为`-`。
- 最终拼成`<plugin-id>-<page-path>`。

例如：

| 文件路径 | 推导结果 |
|------|------|
| `frontend/pages/sidebar-entry.vue` | `plugin-demo-sidebar-entry` |
| `frontend/pages/user/profile.vue` | `plugin-demo-user-profile` |

### 页面开发约束

- 插件页面必须是**真实 Vue 源码文件**，而不是 JSON 描述。
- 页面内容应使用宿主已经公开的前端能力，不要直接依赖宿主未发布的内部实现。
- 如果页面需要请求插件自己的后端接口，建议接口路径保持清晰命名，例如`/plugins/<plugin-id>/summary`。

## 前端 `Slot` 接入

### 目录约定

插件`Slot`统一放在：

```text
frontend/slots/**/*.vue
```

推荐目录结构是“目录名即`slotKey`”，例如：

```text
frontend/slots/
  dashboard.workspace.after/
    workspace-card.vue
```

### `pluginSlotMeta`

`Slot`文件可以导出`pluginSlotMeta`：

```vue
<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceAfter,
};
</script>
```

当前支持字段如下：

| 字段 | 是否必填 | 说明 |
|------|------|------|
| `pluginId` | 否 | 不传时默认从文件路径推导 |
| `slotKey` | 否 | 不传时默认从文件所在目录推导 |
| `order` | 否 | 同一`Slot`下的排序值，越小越靠前，默认`0` |

### 默认 `slotKey` 推导规则

如果文件没有显式声明`slotKey`，宿主会读取其相对路径：

- 先去掉`frontend/slots/`前缀。
- 再去掉文件名。
- 剩余目录路径作为`slotKey`。

例如：

| 文件路径 | 推导出的`slotKey` |
|------|------|
| `frontend/slots/dashboard.workspace.after/workspace-card.vue` | `dashboard.workspace.after` |
| `frontend/slots/auth.login.after/login-tip.vue` | `auth.login.after` |

### 未发布插槽的处理方式

宿主只允许挂载已发布的`slotKey`。如果插件声明了未发布的插槽：

- 宿主会跳过该文件的挂载。
- 控制台会打印告警信息。
- 不会因为单个错误`Slot`影响其他页面或其他`Slot`。

### 当前已发布的前端插槽

| `slotKey` | 宿主位置 | 推荐用途 |
|------|------|------|
| `auth.login.after` | 登录页表单下方 | 提示信息、轻量入口 |
| `crud.table.after` | 通用表格区域下方 | 说明卡片、辅助面板 |
| `crud.toolbar.after` | 通用工具栏右侧 | 状态标签、快捷操作 |
| `dashboard.workspace.before` | 工作台顶部 | 横幅、提醒、概览块 |
| `dashboard.workspace.after` | 工作台底部 | 卡片、统计块、快捷入口 |
| `layout.header.actions.before` | 头部动作区前置 | 全局状态、入口 |
| `layout.header.actions.after` | 头部动作区后置 | 快捷入口、轻量提示 |
| `layout.user-dropdown.after` | 用户菜单左侧 | 轻量入口、状态提示 |

## SQL 约定

### 安装 SQL

插件安装 SQL 放在：

```text
manifest/sql/*.sql
```

规则如下：

| 规则 | 说明 |
|------|------|
| 文件名格式 | 必须是`{序号}-{当前迭代名称}.sql` |
| 序号格式 | 三位数字，例如`001`、`002` |
| 目录层级 | 必须直接位于`manifest/sql/`根目录，不能再嵌套子目录 |
| 扫描顺序 | 宿主按文件名排序后顺序执行 |

### 卸载 SQL

插件卸载 SQL 放在：

```text
manifest/sql/uninstall/*.sql
```

规则如下：

| 规则 | 说明 |
|------|------|
| 文件名格式 | 与安装 SQL 相同 |
| 目录层级 | 必须直接位于`manifest/sql/uninstall/`根目录 |
| 发现方式 | 宿主在卸载流程中按目录约定单独发现 |
| 初始化隔离 | 宿主初始化流程不会扫描该目录，避免误执行卸载 SQL |

### 菜单与权限治理

菜单和权限相关信息必须遵循以下规则：

- 菜单、按钮权限、授权种子统一写在 SQL 中。
- 菜单稳定标识统一使用`sys_menu.menu_key`。
- 菜单父子关系应通过父级`menu_key`解析真实`parent_id`。
- 不要在 SQL 中写死整型`id`或`parent_id`。
- 不要在`plugin.yaml`中再声明菜单结构、权限前缀或菜单前缀。

换句话说：

- 菜单是否存在，以 SQL 为准。
- 插件是否启用，以插件治理状态为准。
- 页面文件是否可挂载，以前端源码文件和宿主运行时为准。

三者各自负责自己的真相源，不互相重复描述。

## 开发步骤

新增一个源码插件时，建议按以下顺序进行：

### 创建插件目录和模块

1. 在`apps/lina-plugins/`下创建`<plugin-id>/`目录。
2. 新建插件自己的`go.mod`。
3. 在根目录`go.work`中加入该插件模块路径。

### 编写最小清单

1. 新建`plugin.yaml`。
2. 只填写最小元数据。
3. 确认`id`使用`kebab-case`，并且与目录语义一致。

### 编写后端入口

1. 新建`backend/plugin.go`。
2. 调用`pluginhost.NewSourcePlugin("<plugin-id>")`。
3. 注册所需的路由、Hook 或其他扩展点。

### 更新宿主显式注册表

1. 修改`apps/lina-plugins/lina-plugins.go`。
2. 新增插件后端包的匿名导入。

### 编写前端页面和`Slot`

1. 页面放到`frontend/pages/`。
2. `Slot`放到`frontend/slots/`。
3. 需要显式元数据时分别导出`pluginPageMeta`和`pluginSlotMeta`。

### 编写 SQL

1. 安装 SQL 放到`manifest/sql/`。
2. 卸载 SQL 放到`manifest/sql/uninstall/`。
3. 菜单和权限写进 SQL，不要再写入`plugin.yaml`。

### 验证

建议至少执行以下验证：

- `go test ./internal/service/plugin ./pkg/pluginhost`
- 插件相关的 E2E 用例
- 手工检查插件管理页、菜单显示、路由访问和禁用后的隐藏行为

## 开发约束

### 后端约束

- 插件后端代码应遵循宿主当前的`GoFrame`目录风格。
- `api/`和`internal/controller/`建议保持与宿主`gf gen ctrl`生成风格一致。
- 公开类型、结构体字段和方法应有足够英文注释，便于人工 review。
- 不要在插件里直接硬编码宿主未公开的内部包路径。

### 前端约束

- 页面和`Slot`必须是可直接参与宿主构建的真实 Vue 文件。
- 优先复用宿主已公开的组件和运行时能力。
- 不要依赖已经被删除的`pages.json`、`slots.json`或类似声明式文件。

### 元数据约束

- 只保留基础元数据。
- 不要把“约定可推导”的信息塞回`plugin.yaml`。
- 不要为了“配置更全”而重建已经被设计移除的模型。

## Review 清单

人工 review 一个源码插件时，建议按下面清单逐项确认：

| 检查项 | 结论标准 |
|------|------|
| 插件目录位置是否正确 | 位于`apps/lina-plugins/<plugin-id>/` |
| 是否存在`go.mod`和`backend/plugin.go` | `source`插件必须具备 |
| `plugin.yaml`是否最小化 | 不应再出现`schemaVersion`、`compatibility`、`entry`、`resources`、`metadata`等字段 |
| `id`是否唯一且符合`kebab-case` | 宿主范围内唯一 |
| `lina-plugins.go`是否补了匿名导入 | 新插件必须显式接线 |
| 页面和`Slot`是否位于约定目录 | 页面在`frontend/pages/`，`Slot`在`frontend/slots/` |
| 菜单和权限是否只在 SQL 中维护 | 不在`plugin.yaml`重复建模 |
| SQL 文件名和目录是否正确 | 安装和卸载 SQL 分别放在正确目录，且文件名合规 |
| 禁用后是否能正确隐藏 | 菜单、页面、`Slot`和路由都应受启停状态保护 |
| 文档是否足够清晰 | 插件自身`README.md`应说明功能范围、路由、SQL 和验证方式 |

## 常见错误

### 插件已写好，但插件管理页看不到

优先检查：

- `plugin.yaml`是否存在。
- `plugin.yaml`字段是否缺失。
- `id`是否与其他插件重复。

### 后端代码编译不过

优先检查：

- 是否创建了插件自己的`go.mod`。
- 根目录`go.work`是否已经包含该插件模块。
- `apps/lina-plugins/lina-plugins.go`是否已经追加匿名导入。

### 页面文件存在，但页面没有挂载

优先检查：

- 文件是否在`frontend/pages/`下。
- 组件是否存在默认导出。
- `pluginPageMeta.routePath`是否与菜单配置对应。

### `Slot`文件存在，但没有渲染

优先检查：

- 文件是否在`frontend/slots/`下。
- `slotKey`是否为宿主已发布的插槽。
- 插件当前是否已启用。

### 菜单存在，但访问返回 404

优先检查：

- 插件是否已启用。
- 后端路由是否通过`RegisterRoutes`正确注册。
- 路由是否挂到了期望的前缀下。

## 参考实现

当前仓库中最小可运行样例是`plugin-demo`：

| 文件 | 作用 |
|------|------|
| `apps/lina-plugins/plugin-demo/plugin.yaml` | 最小清单示例 |
| `apps/lina-plugins/plugin-demo/backend/plugin.go` | 后端注册入口示例 |
| `apps/lina-plugins/plugin-demo/frontend/pages/sidebar-entry.vue` | 插件页面示例 |
| `apps/lina-plugins/plugin-demo/frontend/slots/` | 插件`Slot`示例 |
| `apps/lina-plugins/plugin-demo/manifest/sql/001-plugin-demo.sql` | 菜单与权限种子示例 |
| `apps/lina-plugins/plugin-demo/README.md` | 插件自身接入说明 |

如果要新增新插件，建议先复制`plugin-demo`的整体结构，再按本文档约束删减或扩展，而不是从零随意拼目录。
