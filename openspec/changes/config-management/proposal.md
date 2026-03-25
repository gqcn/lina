## Why

系统需要一个统一的参数设置功能，允许管理员在后台界面中配置和管理影响系统运行时行为的键值对参数。目前系统缺少运行时可配置的参数管理能力，所有配置都依赖配置文件，修改后需要重启服务才能生效。通过数据库存储的参数设置，可以实现运行时动态调整系统行为，无需重启服务。

## What Changes

- 新增 `sys_config` 数据表，存储系统参数（键值对形式）
- 新增后端参数设置 CRUD API（列表查询、详情、新增、修改、删除、按键名查询、导出）
- 新增前端参数设置管理页面，包含搜索栏、数据表格、新增/编辑弹窗
- 新增前端路由和菜单项，将参数设置挂载到系统管理模块下
- 新增参数设置相关的菜单权限和按钮权限数据

## Capabilities

### New Capabilities

- `config-management`: 系统参数设置的完整 CRUD 管理能力，包括列表查询（分页、筛选）、新增、编辑、删除（单条/批量）、按键名查询、Excel 导出

### Modified Capabilities

（无）

## Impact

- **数据库**: 新增 `sys_config` 表，新增菜单和按钮权限种子数据
- **后端 API**: 新增 `/config` 相关 RESTful 接口（7 个端点）
- **前端**: 新增 `src/views/system/config/` 页面、`src/api/system/config/` API 层、路由配置
- **依赖**: 无新增外部依赖，复用现有的 GoFrame + VXE-Grid + Ant Design Vue 技术栈
