# plugin-demo

`plugin-demo` 是一期源码插件 MVP 的示例插件。

本插件强调一条边界：

- 插件特定的前端页面放在 `frontend/`
- 插件特定的后端实现与资源声明放在 `backend/`
- 宿主只提供通用的扫描、启用/禁用、菜单治理、资源查询和 Hook 执行框架

当前目录说明：

```text
  plugin-demo/
  plugin.yaml
  backend/
    plugin.go
  manifest/
    sql/
      uninstall/
  frontend/
    src/pages/*.vue
    src/slots/**/*.vue
```

当前后端能力已支持把插件目录内的 Go 源码与宿主一起编译；`plugin-demo` 通过 `backend/plugin.go` 在编译期注册自己的 Hook 与资源能力，`plugin.yaml` 仅保留插件元数据、入口描述和资源索引，不再声明后端 API 列表。当前前端能力也已支持把插件目录内的 Vue 页面与 Slot 源码纳入宿主构建，由宿主在运行时装载。这是一阶段源码插件的最小接入方式，后续再演进到更完整的 runtime `package/wasm` 机制。

插件插槽目录、类型化 Hook/Slot 常量与推荐接入方式，请优先参考宿主开发指南：`apps/lina-plugins/README.md`。

当前 SQL 约定：

- 安装 SQL 放在 `manifest/sql/` 根目录，并使用 `{序号}-{当前迭代名称}.sql` 命名
- 卸载 SQL 放在 `manifest/sql/uninstall/`，避免被宿主初始化流程误执行
- 插件菜单安装/卸载以 `manifest/sql/` 下的 SQL 为唯一来源
- 菜单稳定标识使用 `menu_key`，父子关系通过父菜单 `menu_key` 解析，不再写死整型 `id`
- 当前仅保留左侧主菜单顶部入口“插件示例”，用于验证源码插件页面可被宿主菜单挂载
- 工作台卡片继续作为前端 Slot 示例保留，用于验证插件启用/禁用后的宿主扩展联动
- 登录审计 Hook 与资源接口继续保留在后端，由宿主通用插件资源接口统一暴露，用于验证插件后端能力，但不再提供独立前端页面入口
