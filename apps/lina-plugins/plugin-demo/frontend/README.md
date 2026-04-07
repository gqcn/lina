# plugin-demo frontend

一期源码插件前端改为提供真实前端文件，供宿主运行时装载：

- `iframe`：可将独立页面地址挂入菜单
- `new-tab`：点击菜单后新标签页打开插件应用
- `host-embed`：宿主内嵌挂载插件页面/微前端

## 建议约定

- 路由前缀：`/plugin-demo-*`
- 权限前缀：`plugin-demo:*`

## 当前实现

- 已提供 `frontend/src/pages/*.vue` 作为插件真实页面源码，当前仅保留左侧菜单示例页
- 已提供 `frontend/src/slots/**/*.vue` 作为工作台扩展的真实 Slot 源码
- 当前一期真正生效的是 `frontend/src/pages/*.vue + frontend/src/slots/**/*.vue + system/plugin/runtime-page` 这条源码挂载链路

## 当前限制

- 当前示例优先验证“插件目录中的前端源码文件可被宿主发现、挂载并以内页 Tab 打开”
- 若插件需要与宿主交互，当前优先通过后端 API 完成；更完整的宿主 SDK / 微前端挂载协议会在后续阶段补齐
