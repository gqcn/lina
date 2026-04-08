# plugin-demo

`plugin-demo`是当前`plugin-framework`迭代的一期源码插件样例，用来验证“插件目录内维护实现 + 宿主侧手工注册 + 前后端按约定发现”的最小闭环。

## 目录结构

```text
plugin-demo/
  go.mod
  plugin.yaml
  README.md
  backend/
    plugin.go
    api/
    internal/controller/
    service/
  frontend/
    pages/
      sidebar-entry.vue
    slots/
      auth.login.after/
      crud.toolbar.after/
      dashboard.workspace.before/
      dashboard.workspace.after/
      layout.header.actions.before/
      layout.header.actions.after/
  manifest/
    sql/
      001-plugin-demo.sql
      uninstall/
        001-plugin-demo.sql
```

## 清单约定

`plugin-demo/plugin.yaml`当前只保留基础元数据：

- `id`
- `name`
- `version`
- `type`
- `description`
- `author`
- `homepage`
- `license`

插件页面、`Slot`、SQL 文件和菜单前缀都不再写入清单，而是分别通过目录约定、页面源码、`Slot`源码和 SQL 本身维护。

## 后端接入

`backend/plugin.go`是当前插件后端接入入口，职责保持单一：

1. 创建`pluginhost.NewSourcePlugin("plugin-demo")`
2. 注册插件后端路由和其他宿主扩展点
3. 由宿主侧[apps/lina-plugins/lina-plugins.go](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/lina-plugins.go)手工匿名导入该插件后端包

当前示例保留两条路由：

| 路由 | 类型 | 说明 |
|------|------|------|
| `GET /api/v1/plugins/plugin-demo/ping` | 匿名访问 | 验证插件可注册公开路由 |
| `GET /api/v1/plugins/plugin-demo/summary` | 鉴权访问 | 验证插件页面可以从插件后端接口取数 |

## 前端接入

当前前端接入完全按目录约定发现：

- `frontend/pages/sidebar-entry.vue` 作为插件页面示例
- `frontend/slots/` 下多个`.vue`文件作为插件`Slot`示例

宿主在构建时会扫描：

- `frontend/pages/**/*.vue`
- `frontend/slots/**/*.vue`

不再要求这些文件在`plugin.yaml`中重复登记。

## SQL 约定

当前 SQL 也按目录约定处理：

- 安装 SQL 位于 `manifest/sql/`
- 卸载 SQL 位于 `manifest/sql/uninstall/`

菜单相关的`menu_key`、权限码和父子关系都以 SQL 为单一真相源，不再在清单中重复声明。

## Review 关注点

人工 review `plugin-demo`时，建议重点核对：

| 检查项 | 说明 |
|------|------|
| `plugin.yaml`是否保持最小化 | 不应再出现`schemaVersion`、`compatibility`、`resources`、`metadata`等重复配置 |
| 宿主注册是否显式 | `apps/lina-plugins/lina-plugins.go`中应手工导入插件后端包 |
| 页面与 `Slot` 是否按目录约定发现 | 不依赖清单额外声明 |
| 菜单和权限是否只在 SQL 中维护 | 不在元数据中重复建模 |
