# Lina Plugins 开发指南

`apps/lina-plugins/`用于承载`Lina`插件源码与插件开发文档。本文档面向插件开发者，说明源码插件的目录约定、前后端插槽目录，以及推荐的类型化接入方式。

## 目录结构

当前源码插件统一放在`apps/lina-plugins/<plugin-id>/`下，推荐结构如下：

```text
apps/lina-plugins/
  README.md
  <plugin-id>/
    plugin.yaml
    backend/
      plugin.go
    frontend/
      src/pages/*.vue
      src/slots/**/*.vue
    manifest/
      sql/
        001-<plugin-name>.sql
        uninstall/
          001-<plugin-name>.sql
```

关键约束如下：

| 项目 | 约束                             |
|------|--------------------------------|
| `plugin.yaml` | 维护插件元数据、资源索引与前端接入提示            |
| `backend/plugin.go` | 使用`Go`代码通过回调注册源码插件的后端扩展点与资源能力    |
| `frontend/pages/` | 提供插件页面源码，交由宿主运行时页装载                |
| `frontend/slots/` | 提供插件`Slot`源码，交由宿主公开扩展点装载             |
| `manifest/sql/` | 存放安装`SQL`，命名遵循`{序号}-{当前迭代名称}.sql` |
| `manifest/sql/uninstall/` | 存放卸载`SQL`，避免被宿主初始化流程误扫           |

## 扩展点模型

`Lina`中的“插件可安装位置”分为两类：

| 类别 | 含义 | 类型定义位置 |
|------|------|-------------|
| 后端扩展点 | 宿主主服务上公开的类型化回调注册点 | `apps/lina-core/pkg/pluginhost/pluginhost_slots.go` |
| 前端扩展点 | 宿主公开`UI`容器上的内容插入点 | `apps/lina-vben/apps/web-antd/src/plugins/plugin-slots.ts` |

插件开发时必须直接引用这些类型定义，不能在插件代码里硬编码扩展点字符串。

## 后端扩展点

当前源码插件后端以**回调注册**为统一模型。事件型 `Hook` 与路由/鉴权后回调/定时任务/过滤器，本质上都属于“插件向宿主某个已发布后端扩展点注册回调函数”：

- 扩展点常量统一使用 `pluginhost.ExtensionPoint*`
- 回调执行模式统一使用 `pluginhost.CallbackExecutionMode*`
- 插件和宿主都不应再硬编码 `auth.login.succeeded`、`http.route.register` 这类裸字符串
- 对插件开发者而言，只需要理解“选择一个 `ExtensionPoint`，再注册一个带执行模式的回调函数”；事件触发和注册触发只是宿主内部的触发语义区别

### 1. 已发布后端扩展点目录

| Go 常量 | Canonical 值 | 触发语义 | 支持模式 | 常见用途 |
|------|------|------|------|------|
| `ExtensionPointAuthLoginSucceeded` | `auth.login.succeeded` | 事件触发 | `blocking`,`async` | 登录审计、登录后初始化 |
| `ExtensionPointAuthLoginFailed` | `auth.login.failed` | 事件触发 | `blocking`,`async` | 失败审计、风控记录 |
| `ExtensionPointAuthLogoutSucceeded` | `auth.logout.succeeded` | 事件触发 | `blocking`,`async` | 会话回收、退出审计 |
| `ExtensionPointSystemStarted` | `system.started` | 事件触发 | `blocking`,`async` | 启动预热、自检、注册 |
| `ExtensionPointPluginInstalled` | `plugin.installed` | 事件触发 | `blocking`,`async` | 插件自初始化 |
| `ExtensionPointPluginEnabled` | `plugin.enabled` | 事件触发 | `blocking`,`async` | 启用后补偿逻辑 |
| `ExtensionPointPluginDisabled` | `plugin.disabled` | 事件触发 | `blocking`,`async` | 禁用后清理逻辑 |
| `ExtensionPointPluginUninstalled` | `plugin.uninstalled` | 事件触发 | `blocking`,`async` | 卸载回收逻辑 |
| `ExtensionPointHTTPRouteRegister` | `http.route.register` | 注册触发 | `blocking` | 注册插件自有 API |
| `ExtensionPointHTTPRequestAfterAuth` | `http.request.after-auth` | 注册触发 | `blocking` | 鉴权后补充响应头、请求上下文处理 |
| `ExtensionPointCronRegister` | `cron.register` | 注册触发 | `blocking` | 注册插件定时任务 |
| `ExtensionPointMenuFilter` | `menu.filter` | 注册触发 | `blocking` | 对宿主菜单进行可见性过滤 |
| `ExtensionPointPermissionFilter` | `permission.filter` | 注册触发 | `blocking` | 对宿主权限进行可见性过滤 |

说明：

1. `blocking` 表示回调在宿主当前流程内执行，可以阻塞宿主后续步骤。
2. `async` 表示宿主将回调异步执行，插件回调不能再假设自己能阻塞当前主流程。
3. 当前只有“事件触发”类后端扩展点支持 `async`；“注册触发”类点位出于一致性和正确性要求，仅支持 `blocking`。
4. 宿主传入插件的 `HookPayload`、`AfterAuthInput`、`RouteRegistrar`、`CronRegistrar`、`MenuDescriptor`、`PermissionDescriptor` 都是接口对象，而不是宿主内部结构体指针。
5. `CronRegistrar` 额外暴露 `IsPrimaryNode()`，插件可据此决定某些定时逻辑是否只在主节点执行。
6. 插件 HTTP 路由注册使用宿主单独开放的无前缀插件路由根分组；插件可通过 `RouteRegistrar.Group(prefix, func(group *ghttp.RouterGroup) { ... })` 自行决定是否使用 `/api/v1` 等前缀。
7. `RouteRegistrar.Middlewares()` 会公开宿主已发布的中间件目录；插件可按需组合 `NeverDoneCtx`、`HandlerResponse`、`CORS`、`Ctx`、`Auth`、`OperLog`，也可以与插件自定义中间件混用。
8. 若同一插件需要同时暴露免鉴权和需鉴权接口，直接通过 `RouteRegistrar.Group(prefix, func(group *ghttp.RouterGroup) { ... })` 创建外层分组，再在组内按需拆分子分组和组合宿主中间件。

### 2. 推荐注册方式

```go
package backend

import "lina-core/pkg/pluginhost"

func init() {
	plugin := pluginhost.NewSourcePlugin("plugin-demo")
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuthLoginSucceeded,
		pluginhost.CallbackExecutionModeBlocking,
		writeAuditEvent,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuthLoginFailed,
		pluginhost.CallbackExecutionModeAsync,
		writeAuditEvent,
	)
	plugin.RegisterAfterAuthHandler(
		pluginhost.ExtensionPointHTTPRequestAfterAuth,
		pluginhost.CallbackExecutionModeBlocking,
		markAfterAuthRequest,
	)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerCrons,
	)
	plugin.RegisterResource(&pluginhost.ResourceSpec{
		Key: "example-records",
	})
	pluginhost.RegisterSourcePlugin(plugin)
}
```

```go
func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	if registrar == nil {
		return nil
	}

	middlewares := registrar.Middlewares()
	if middlewares == nil {
		return nil
	}

	demoController := democtrl.NewV1()
	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(demoController.Ping)
		})

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.OperLog(),
			)
			group.Bind(demoController.Summary)
		})
	})
	return nil
}
```

约束如下：

1. 事件 `Hook`、路由、定时任务、过滤器都通过回调注册，并显式声明执行模式。
2. 宿主传入插件的对象型参数统一使用接口抽象，插件只依赖宿主公开方法，不依赖宿主内部结构体。
3. 若插件需要只在主节点执行业务，可在 `CronRegistrar` 回调中通过 `IsPrimaryNode()` 做分支判断。
4. 插件自有查询数据仍建议通过 `ResourceSpec` 暴露给宿主统一资源 API。
5. 插件注册未知后端扩展点，或为某扩展点声明不支持的执行模式时，宿主会在注册阶段拒绝该声明。
6. 宿主不再为插件 HTTP 路由隐式附加 `/api/v1` 等前缀；插件需自行声明目标路由前缀。
7. `Group()` 回调中可继续使用原生 `group.Group()` 拆分子分组，因此同一插件可以像宿主主服务一样组织公开/受保护路由，并按需组合宿主中间件。
8. 插件后端 `api/` 与 `controller/` 目录也必须遵循宿主 GoFrame 脚手架规范：`api/<module>/<module>.go + api/<module>/v1/*.go`，以及 `internal/controller/<module>/<module>.go + <module>_new.go + <module>_v1_*.go` 这类 `gf gen ctrl` 风格命名。

## 前端扩展点

以下前端 `UI` 扩展点已正式发布，可在源码插件前端中使用：

| 扩展点 | 宿主位置 | 推荐内容 |
|------|---------|---------|
| `auth.login.after` | 登录页表单下方 | 提示信息、轻量入口 |
| `layout.header.actions.before` | 头部动作区前置 | 全局状态、全局入口 |
| `layout.header.actions.after` | 头部动作区后置 | 快捷入口、状态提示 |
| `layout.user-dropdown.after` | 后台右上角用户菜单左侧 | 轻量入口、状态提示、快捷操作 |
| `dashboard.workspace.before` | 工作台主内容区顶部 | 横幅、概览块、提醒 |
| `dashboard.workspace.after` | 工作台主内容区底部 | 卡片、统计块、快捷入口 |
| `crud.toolbar.after` | 通用`CRUD`工具栏右侧 | 状态标签、快捷操作 |
| `crud.table.after` | 通用`CRUD`表格区域下方 | 说明卡片、辅助面板 |

前端源码插件应当从宿主前端导出的扩展点常量中引用合法位置：

```vue
<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceAfter,
};
</script>
```

宿主页面与插件装载器也使用同一份常量，因此插件一旦声明未发布的扩展点，宿主会跳过装载并记录错误。

## 开发约束

开发插件时请遵循以下规则：

1. 不要在插件代码中直接写`auth.login.succeeded`、`http.route.register`、`crud.toolbar.after`这类裸字符串，统一引用宿主定义的类型常量。
2. 只使用本文档“已发布的扩展点”，不要假设宿主存在未文档化的私有扩展点或回调。
3. 插件页面源码放在`frontend/pages/`，插件Slot源码放在`frontend/slots/`，不要混放。
4. 插件SQL中的菜单与权限仍然通过宿主治理体系接入，菜单稳定标识使用`menu_key`，不要写死整型`id`。
5. 若需要新增扩展点，必须先更新`OpenSpec`规格、宿主类型定义与本文档，再开始实现插件接入。

## 示例参考

当前仓库提供了`plugin-demo`作为最小源码插件样板，可直接参考以下文件：

| 文件 | 作用 |
|------|------|
| `apps/lina-plugins/plugin-demo/plugin.yaml` | 插件元数据与资源索引 |
| `apps/lina-plugins/plugin-demo/backend/plugin.go` | 后端`Route`、`After-Auth`、`Cron`与接口化回调注册示例 |
| `apps/lina-plugins/plugin-demo/frontend/pages/sidebar-entry.vue` | 左侧菜单页面示例 |
| `apps/lina-plugins/plugin-demo/frontend/slots/` | 多个前端`Slot`示例 |

如果需要新增插件，建议先复制`plugin-demo`目录结构，再按本文档调整插件`ID`、`SQL`、页面与回调注册代码。
