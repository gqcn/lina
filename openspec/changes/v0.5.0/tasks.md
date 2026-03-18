## 1. 后端：系统信息 API

- [x] 1.1 创建 SQL 文件 `manifest/sql/v0.5.0.sql`（本版本无 DDL 变更，仅作为版本占位）
- [x] 1.2 创建 API 定义 `api/sysinfo/v1/info.go`，定义 `GET /api/v1/system/info` 的请求/响应结构体（返回 Go 版本、GoFrame 版本、OS、数据库版本、启动时间、运行时长）
- [x] 1.3 执行 `make ctrl` 生成控制器骨架
- [x] 1.4 实现 `internal/service/sysinfo/sysinfo.go` 系统信息服务层，获取运行时信息
- [x] 1.5 在控制器中调用服务层并返回数据
- [x] 1.6 在 `internal/cmd/cmd_http.go` 中注册系统信息路由（鉴权路由组内）

## 2. 前端：路由与菜单结构

- [x] 2.1 创建路由模块 `src/router/routes/modules/about.ts`，定义"系统信息"顶级菜单及三个子路由（系统接口、系统信息、组件演示）
- [x] 2.2 创建视图目录 `src/views/about/`，包含 `api-docs/index.vue`、`system-info/index.vue`、`component-demo/index.vue` 三个页面组件骨架

## 3. 前端：系统接口页面（Scalar OpenAPI UI）

- [x] 3.1 安装 `@scalar/api-reference` npm 依赖
- [x] 3.2 实现 `src/views/about/api-docs/index.vue`，集成 Scalar Vue 组件，加载后端 `/api.json`
- [x] 3.3 创建前端配置文件 `src/views/about/config.ts`，定义 OpenAPI 规范地址、组件演示地址、项目信息、后端/前端组件列表等可配置项

## 4. 前端：系统信息页面

- [x] 4.1 创建 API 文件 `src/api/about/index.ts`，定义 `getSystemInfo` 方法调用 `GET /api/v1/system/info`
- [x] 4.2 实现 `src/views/about/system-info/index.vue`，包含四个 Card 区块：关于项目、基本信息、后端组件、前端组件
- [x] 4.3 关于项目区块：从配置对象读取项目名称、描述、版本、许可证、主页链接
- [x] 4.4 基本信息区块：调用后端 API 展示运行时数据（Go 版本、OS、数据库版本、启动时间、运行时长等）
- [x] 4.5 后端/前端组件区块：从配置对象读取组件列表，以网格布局展示名称、版本、可点击外链

## 5. ~~前端：组件演示页面~~（已取消）

- [x] ~~5.1 实现组件演示页面~~（已取消）
- [x] ~~5.2 实现 iframe 加载失败检测~~（已取消）

## 6. E2E 测试

- [x] 6.1 创建 `hack/tests/e2e/about/` 测试目录
- [x] 6.2 编写 `TC0044-api-docs-page.ts`：验证系统接口页面加载 Scalar UI 正常展示
- [x] 6.3 编写 `TC0045-system-info-page.ts`：验证系统信息页面四个区块正常展示，后端数据正确加载
- [x] ~~6.4 编写 `TC0046-component-demo-page.ts`~~（已取消）
- [x] 6.5 运行全部 E2E 测试确认无回归（114 passed，6 failed 均为已有问题，新增 3 个测试全部通过）

## Feedback

- [x] **FB-1**：~~系统接口页面 Scalar API Client 弹窗被遮挡~~ → 已改用 Stoplight Elements 替代 Scalar，通过 Web Component 方式集成，样式完全隔离无冲突
- [x] **FB-2**：系统接口页面顶部空白过多，需减少顶部间距
- [x] **FB-3**：系统接口页面左侧 Overview 菜单点击后右侧内容为空白
- [x] **FB-4**：系统接口页面左侧接口分类菜单应粗体展示
- [x] **FB-5**：系统接口页面 Stoplight CSS 污染全局页面样式（边框消失等），改用 iframe 嵌入实现样式隔离
- [x] **FB-6**：系统接口文档 HTML 页面应改为静态文件方式提供，移除后端 API 路由，减少系统复杂度
