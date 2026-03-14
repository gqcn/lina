# CLAUDE.md

## 项目概述

Lina 是一个管理后台系统，采用前后端分离架构。

- **前端**: Vben5 + Vue 3 + Ant Design Vue + TypeScript（pnpm monorepo）
- **后端**: GoFrame v2 + SQLite + JWT
- **参考项目**: `/Users/john/Workspace/gitee/dapppp/ruoyi-plus-vben5`（前端样式和功能参考）

## 目录结构

```
apps/
  backend/           → GoFrame v2 后端
    api/             → 请求/响应 DTO（g.Meta 路由定义）
    internal/
      cmd/           → 服务启动 & 路由注册
      consts/        → 常量定义
      controller/    → HTTP 控制器（gf gen ctrl 自动生成骨架）
      dao/           → 数据访问层（gf gen dao 自动生成）
      model/         → 数据模型
        do/          → 数据操作对象（自动生成）
        entity/      → 数据库实体（自动生成）
      service/       → 业务逻辑层
    manifest/
      config/        → 配置文件
      data/          → SQLite 数据库文件
      sql/           → DDL + Seed DML（版本 SQL 文件）
        mock-data/   → Mock 演示/测试数据（不随生产部署）
  frontend/          → Vben5 前端（pnpm monorepo）
    apps/web-antd/   → 主应用（Ant Design Vue）
    packages/        → 共享库（@core, effects, stores, utils 等）
hack/
  tests/             → E2E 测试（Playwright）
    e2e/             → 测试用例文件
    fixtures/        → 测试 fixtures（auth, config）
    pages/           → 页面对象模型
openspec/
  changes/           → OpenSpec 变更记录
```

## 常用命令

### 开发环境

```bash
make dev       # 启动前后端（前端:5666, 后端:8080）
make stop      # 停止所有服务
make status    # 查看服务状态
make test      # 运行 E2E 测试
make init      # 初始化数据库（DDL + Seed 数据）
make mock      # 加载 Mock 演示数据（需先执行 init）
make up        # AI 生成 commit message 并推送
```

### 后端

```bash
cd apps/backend
go run main.go          # 运行
make build              # 构建
make dao                # 生成 DAO/DO/Entity（修改 SQL 后）
make ctrl               # 生成控制器骨架（修改 API 定义后）
```

### 前端

```bash
cd apps/frontend
pnpm install                    # 安装依赖
pnpm -F @lina/web-antd dev     # 开发模式
pnpm run build                  # 构建
```

### E2E 测试

```bash
cd hack/tests
pnpm test              # 运行全部测试
pnpm test:headed       # 带浏览器界面运行
pnpm test:ui           # 交互式测试界面
pnpm test:debug        # 调试模式
pnpm report            # 查看 HTML 报告
```

测试文件命名规范：`TC{NNNN}*.ts`（如 `TC0001-login.ts`），放在 `hack/tests/e2e/` 对应模块目录下。

## 开发规范

### 后端

- 遵循 GoFrame v2 标准分层架构
- DAO/DO/Entity 由 `gf gen dao` 自动生成，**不要手动修改**
- Controller 由 `gf gen ctrl` 自动生成骨架，在生成的文件中填写业务逻辑
- 数据库操作**必须使用 DO 对象**，不使用 `g.Map`
- 错误处理使用 `gerror` 组件
- **优先使用 GoFrame 内置方法**：所有 Go 方法调用优先使用 GoFrame 框架已有的方法，避免重复造轮子。例如遍历目录使用 `gfile.ScanDirFile`，而非自行实现目录遍历逻辑
- **SQL 文件按迭代版本命名**：每次数据库变更的 SQL 文件以当前迭代版本号命名（如 `v0.1.0.sql`、`v0.2.0.sql`），存放在 `manifest/sql/` 目录下。`init.sql` 仅用于初始建表，后续迭代的表结构变更（ALTER TABLE、新增表等）使用版本号命名的 SQL 文件。升级时按版本顺序依次执行即可完成数据库迁移。
- **SQL 数据分类管理**：版本 SQL 文件（如 `v0.2.0.sql`）中只允许包含 DDL（建表/改表）和 Seed DML（系统运行所必需的初始化数据，如字典类型、管理员账号等）。演示/测试用的 Mock 数据（如测试用户、演示部门/岗位等）必须放到 `manifest/sql/mock-data/` 目录下的独立 SQL 文件中，文件名以数字前缀控制执行顺序（如 `01_mock_depts.sql`、`02_mock_posts.sql`）。
- 代码生成流程：
  1. **API 变更**: 修改 `api/{resource}/v1/*.go` → `make ctrl`
  2. **数据库变更**: 新增或修改 `manifest/sql/{version}.sql`（如 `v0.2.0.sql`）→ `make dao`
  3. **Service 变更**: 手动在 `internal/service/` 中实现

### 前端

- 路径别名 `#/*` 指向 `./src/*`
- 路由模块放 `src/router/routes/modules/`
- 视图文件放 `src/views/` 对应目录
- API 文件放 `src/api/` 对应目录
- 适配器层 `src/adapter/`：component（组件映射）、form（表单配置）、vxe-table（表格配置）
- 全局组件在 `src/components/global/` 注册（如 GhostButton 用于表格操作列）
- 表格页面使用 `useVbenVxeGrid` + `Page` 组件，操作列用 `ghost-button` + `Popconfirm`
- 前端样式和交互参考 ruoyi-plus-vben5 项目保持一致

### E2E 测试要求

- 修复 bug 或新增功能涉及**用户可观察行为变化**时，必须编写或更新对应的 E2E 测试用例
- 修改完成后必须运行相关 E2E 测试并确认通过，再标记任务完成
- 纯内部重构（无 UI 变化）可豁免，但需运行已有测试套件确认无回归

### UI 设计规范

**重要：所有前端 UI 设计和实现必须参考 ruoyi-plus-vben5 项目（`/Users/john/Workspace/gitee/dapppp/ruoyi-plus-vben5`），保持 UI 的一致性和用户体验的一致性。**

在实现任何前端页面或组件时，必须遵循以下规范：

1. **UI 交互设计**: 弹窗（Modal/Drawer）、表单、表格、搜索栏等交互模式必须与参考项目保持一致
2. **页面样式**: 布局、间距、字体、颜色等视觉元素参考参考项目的实现
3. **组件使用**: 优先使用与参考项目相同的组件和配置方式，包括：
   - 表单使用 `useVbenForm`，弹窗使用 `useVbenModal`，抽屉使用 `useVbenDrawer`
   - RadioGroup 单选项使用 `optionType: 'button'` + `buttonStyle: 'solid'`（按钮样式）
   - 文件上传使用 `Upload.Dragger`（拖拽上传样式）
   - 文件下载使用 `requestClient.download` 方法
   - 操作列的"更多"下拉菜单使用 `Dropdown` + `Menu` + `MenuItem`
4. **弹窗规范**: 导入弹窗包含拖拽上传区域、文件类型提示、下载模板链接、覆盖开关；重置密码弹窗包含用户信息展示（Descriptions）和密码输入
5. **图标使用**: 使用 `IconifyIcon` 组件（来自 `@vben/icons`），图标名使用 Iconify 格式（如 `ant-design:inbox-outlined`）

开发新页面前，**必须先查看参考项目中对应页面的实现**，确保 UI 和交互保持一致。

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

## OpenSpec 工作流

本项目采用 OpenSpec 驱动开发。变更记录存放在 `openspec/changes/` 目录下。每个变更包含：`proposal.md`（提案）、`design.md`（设计）、`specs/`（规格说明）、`tasks.md`（任务清单）。

**关键规则**：
1. 当用户报告缺陷/问题/反馈时（无论中文或英文），必须先调用 `openspec-feedback` 技能，**然后再进行任何代码修改**。

## GoFrame v2 开发标准

**重要：所有 Go 后端代码必须遵循 GoFrame v2 框架规范。所有后端开发任务必须使用 `goframe-v2` 技能。**

### 何时使用 goframe-v2 技能

**强制要求** — 以下场景必须调用 `goframe-v2` 技能：
- 编写或修改 `apps/backend/` 中的任何 Go 代码
- 实现服务层逻辑（`internal/service/`）
- 创建新的 API 接口或控制器
- 进行数据库操作（DAO/ORM）
- 实现错误处理、日志记录或数据校验
- 管理配置、上下文或依赖注入
- 任何 Go 后端开发任务

### 服务层要求

服务层代码（`internal/service/`）必须遵循以下模式：

1. **接口定义**：在 `internal/service/{domain}/` 中定义清晰的服务接口方法签名
2. **实现结构**：通过 `s{ServiceName}` 结构体实现服务，支持依赖注入
3. **上下文管理**：第一个参数始终传入 `ctx context.Context`
4. **错误处理**：使用 GoFrame 的 `gerror` 包进行结构化错误处理
5. **日志记录**：使用 `g.Log()` 并传入上下文进行日志记录
6. **配置访问**：使用 `g.Cfg()` 获取配置项
7. **数据校验**：使用 GoFrame 的校验标签和 `gvalid` 包
8. **事务管理**：使用 `g.DB().Transaction()` 处理多步操作
