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

| 项目 | 约束 |
|------|------|
| `plugin.yaml` | 维护插件元数据、资源索引与前端接入提示 |
| `backend/plugin.go` | 使用Go代码注册源码插件的后端Hook与资源能力 |
| `frontend/src/pages/` | 提供插件页面源码，交由宿主运行时页装载 |
| `frontend/src/slots/` | 提供插件Slot源码，交由宿主公开插槽装载 |
| `manifest/sql/` | 存放安装SQL，命名遵循`{序号}-{当前迭代名称}.sql` |
| `manifest/sql/uninstall/` | 存放卸载SQL，避免被宿主初始化流程误扫 |

## 插槽模型

`Lina`中的“插件可安装位置”分为两类：

| 类别 | 含义 | 类型定义位置 |
|------|------|-------------|
| 后端插槽 | 宿主关键业务事件上的Hook扩展点 | `apps/lina-core/pkg/pluginhost/hook_slots.go` |
| 前端插槽 | 宿主公开UI容器上的内容插入点 | `apps/lina-vben/apps/web-antd/src/plugins/plugin-slots.ts` |

插件开发时必须直接引用这些类型定义，不能在插件代码里硬编码插槽字符串。

## 已发布的后端插槽

以下后端Hook插槽已正式发布，可在源码插件后端中使用：

| 插槽 | 触发时机 | 常见用途 |
|------|---------|---------|
| `auth.login.succeeded` | 用户登录成功后 | 登录审计、登录后初始化 |
| `auth.logout.succeeded` | 用户登出成功后 | 会话回收、退出审计 |
| `system.started` | HTTP服务启动完成后 | 启动预热、自检、注册 |
| `plugin.installed` | 运行时插件安装完成后 | 插件自初始化 |
| `plugin.enabled` | 插件启用完成后 | 启用后补偿逻辑 |
| `plugin.disabled` | 插件禁用完成后 | 禁用后清理逻辑 |
| `plugin.uninstalled` | 运行时插件卸载完成后 | 卸载回收逻辑 |

后端Go插件应当从`pluginhost`包引用插槽常量和动作常量：

```go
package backend

import "lina-core/pkg/pluginhost"

func init() {
	pluginhost.RegisterSourcePlugin(&pluginhost.SourcePlugin{
		ID: "plugin-demo",
		Hooks: []*pluginhost.HookSpec{
			{
				Event:  pluginhost.HookSlotAuthLoginSucceeded,
				Action: pluginhost.HookActionInsert,
				Table:  "plugin_demo_login_audit",
				Fields: map[string]string{
					"user_name": "event.userName",
				},
			},
		},
	})
}
```

## 已发布的前端插槽

以下前端UI插槽已正式发布，可在源码插件前端中使用：

| 插槽 | 宿主位置 | 推荐内容 |
|------|---------|---------|
| `layout.user-dropdown.after` | 后台右上角用户菜单左侧 | 轻量入口、状态提示、快捷操作 |
| `dashboard.workspace.after` | 工作台主内容区底部 | 卡片、统计块、快捷入口 |

前端源码插件应当从宿主前端导出的插槽常量中引用合法位置：

```vue
<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceAfter,
};
</script>
```

宿主页面与插件装载器也使用同一份常量，因此插件一旦声明未发布的插槽，宿主会跳过装载并记录错误。

## 后续规划插槽

以下插槽已经进入设计，但当前一期源码插件底座尚未正式发布，插件开发者暂时不要依赖：

| 类别 | 规划插槽 | 说明 |
|------|---------|------|
| 前端 | `auth.login.after` | 登录页公开扩展区 |
| 前端 | `system.user.detail.after` | 用户详情页扩展区 |

后续若新增正式插槽，宿主会同步更新`OpenSpec`与本文档，再对外发布。

## 开发约束

开发插件时请遵循以下规则：

1. 不要在插件代码中直接写`auth.login.succeeded`、`dashboard.workspace.after`这类裸字符串，统一引用宿主定义的类型常量。
2. 只使用本文档“已发布的插槽”，不要依赖尚未发布的规划插槽。
3. 插件页面源码放在`frontend/src/pages/`，插件Slot源码放在`frontend/src/slots/`，不要混放。
4. 插件SQL中的菜单与权限仍然通过宿主治理体系接入，菜单稳定标识使用`menu_key`，不要写死整型`id`。
5. 若需要新增插槽，必须先更新`OpenSpec`规格、宿主类型定义与本文档，再开始实现插件接入。

## 示例参考

当前仓库提供了`plugin-demo`作为最小源码插件样板，可直接参考以下文件：

| 文件 | 作用 |
|------|------|
| `apps/lina-plugins/plugin-demo/plugin.yaml` | 插件元数据与资源索引 |
| `apps/lina-plugins/plugin-demo/backend/plugin.go` | 后端Hook与资源注册示例 |
| `apps/lina-plugins/plugin-demo/frontend/src/pages/sidebar-entry.vue` | 左侧菜单页面示例 |
| `apps/lina-plugins/plugin-demo/frontend/src/slots/dashboard.workspace.after/workspace-card.vue` | 工作台Slot示例 |

如果需要新增插件，建议先复制`plugin-demo`目录结构，再按本文档调整插件ID、SQL、页面与Hook声明。
