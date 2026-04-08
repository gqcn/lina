# plugin-demo

`plugin-demo` 是一期源码插件 MVP 的示例插件。

本插件强调一条边界：

- 插件特定的前端页面放在 `frontend/`
- 插件特定的后端实现与资源声明放在 `backend/`
- 插件元数据与安装资源放在 `plugin.yaml`、`manifest/`
- 宿主只提供通用的扫描、启用/禁用、菜单治理与回调注册式扩展框架

当前目录说明：

```text
plugin-demo/
  plugin.yaml
  backend/
    plugin.go
  frontend/
    pages/
    slots/
  manifest/
    sql/
      001-plugin-demo.sql
      uninstall/
        001-plugin-demo.sql
```

当前后端能力已支持把插件目录内的 Go 源码与宿主一起编译；`plugin-demo` 通过 `backend/plugin.go` 在编译期注册最小化的后端路由回调示例，并统一使用 `pluginhost.ExtensionPoint* + pluginhost.CallbackExecutionMode*` 常量声明宿主安装点与执行模式。`plugin.yaml` 仅保留插件元数据、入口描述和资源索引，不再声明后端 API 列表。当前前端能力也已支持把插件目录内的 Vue 页面与 Slot 源码纳入宿主构建，由宿主在运行时装载。这是一阶段源码插件的最小接入方式，后续再演进到更完整的 runtime `package/wasm` 机制。

插件插槽目录、类型化 Hook/Slot 常量与推荐接入方式，请优先参考宿主开发指南：`apps/lina-plugins/README.md`。

## 后端实现

`backend/` 目录存放 `plugin-demo` 的后端 Go 源码实现。源码插件的后端能力以本目录的 Go 注册代码为准，而不是由 `plugin.yaml` 直接声明后端 API。

- `backend/plugin.go` 在编译期注册插件订阅的宿主 HTTP 路由，是当前最小后端接入示例。
- `backend/api/demo` 与 `backend/internal/controller/demo` 目录命名遵循宿主现有 GoFrame `gf gen ctrl` 约定，保持接口定义和控制器文件命名风格一致。
- 路由示例通过 `pluginhost.ExtensionPointHTTPRouteRegister` 获取宿主开放的无前缀插件路由根分组，并使用与宿主主服务一致的 `group.Group(..., func(group *ghttp.RouterGroup){ ... })` 风格注册：
  - 外层 `registrars.Group("/api/v1", func(group *ghttp.RouterGroup) { ... })` 先挂基础中间件
  - 内层匿名子分组 `group.Group("/", func(group *ghttp.RouterGroup) { group.Bind(demoController.Ping) })` 注册免鉴权 `GET /api/v1/plugins/plugin-demo/ping`
  - 内层鉴权子分组 `group.Group("/", func(group *ghttp.RouterGroup) { group.Middleware(Auth, OperLog); group.Bind(demoController.Summary) })` 注册需鉴权 `GET /api/v1/plugins/plugin-demo/summary`
- 这些分组都受插件启停控制，是否鉴权完全由插件自行选择是否组合宿主公开的 `Auth`、`OperLog` 等中间件决定。
- `RouteRegistrar` 等宿主暴露给插件的回调输入对象均为接口类型，避免插件与宿主内部结构体强耦合。
- `plugin-demo` 只演示最小源码插件接入，不承担数据库读写示例；宿主 `lina-core` 不再手写 `plugin-demo` 专属控制器、服务或路由逻辑。

## 前端实现

`frontend/` 目录提供插件真实前端源码，交由宿主在运行时发现和装载。

- `frontend/pages/` 存放插件页面源码，当前保留左侧菜单示例页，用于验证插件页面可通过宿主菜单与运行时页装载链路挂载。
- `frontend/slots/` 存放插件前端 Slot 源码，当前覆盖登录页、头部动作区、工作台与 CRUD 通用壳层。
- 当前真实生效的接入链路是 `frontend/pages/*.vue + frontend/slots/**/*.vue + system/plugin/runtime-page`。
- 建议约定保持 `route` 前缀为 `/plugin-demo-*`、权限前缀为 `plugin-demo:*`。
- 当前示例优先验证“插件目录中的前端源码文件可被宿主发现、挂载并以内页 Tab 打开”；更完整的宿主 SDK 与微前端挂载协议留待后续阶段。

## 清单与安装资源

`plugin.yaml` 是插件统一入口清单，负责维护插件元数据、入口描述并索引 `manifest/` 下的资源文件，但不负责注册后端路由或后端回调。

- 安装 SQL 放在 `manifest/sql/` 根目录，并使用 `{序号}-{当前迭代名称}.sql` 命名。
- 卸载 SQL 放在 `manifest/sql/uninstall/`，避免被宿主初始化流程误执行。
- 一期源码插件 MVP 下，宿主初始化流程只扫描 `sql/` 根目录，不会顺序执行 `sql/uninstall/`。
- 插件菜单安装与卸载以 `manifest/sql/` 下的 SQL 为单一真相源。
- 菜单稳定标识使用 `sys_menu.menu_key`，父子关系通过父菜单 `menu_key` 解析，不再写死整型 `id`。
- 当前 `plugin-demo` 不再创建插件私有业务表，安装 SQL 只保留菜单与授权种子。
- 若宿主已有对应对象，安装 SQL 应优先使用幂等写法，例如 `INSERT IGNORE`。

## 当前示例覆盖范围

- 左侧主菜单入口“插件示例”，用于验证源码插件页面可被宿主菜单挂载。
- 前端 Slot 示例覆盖 `auth.login.after`、`layout.header.actions.before`、`layout.header.actions.after`、`dashboard.workspace.before`、`dashboard.workspace.after`、`crud.toolbar.after`。
- 后端示例仅保留一个公开 `ping` 路由和一个受保护 `summary` 路由，用于验证最小路由注册式扩展点与页面取数链路。
