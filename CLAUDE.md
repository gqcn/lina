# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
      sql/           → 数据库 schema
  frontend/          → Vben5 前端（pnpm monorepo）
    apps/web-antd/   → 主应用（Ant Design Vue）
    packages/        → 共享库（@core, effects, stores, utils 等）
hack/
  tests/             → E2E 测试（Playwright）
    e2e/             → 测试用例文件
    fixtures/        → 测试 fixtures（auth, config）
    pages/           → Page Object Models
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
- 代码生成流程：
  1. **API 变更**: 修改 `api/{resource}/v1/*.go` → `make ctrl`
  2. **数据库变更**: 修改 `manifest/sql/init.sql` → `make dao`
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

## 默认账号

- 用户名: `admin`
- 密码: `admin123`

## OpenSpec Workflow

This project uses OpenSpec-driven development. Changes live in `openspec/changes/`. Each change has: `proposal.md`, `design.md`, `specs/`, `tasks.md`.

**Critical rules**:
1. When user reports bugs/issues/feedback (Chinese or English), invoke the `openspec-feedback` skill BEFORE making any code changes.

## GoFrame v2 Development Standards

**CRITICAL: All Go backend code MUST follow GoFrame v2 framework conventions. Use the `goframe-v2` skill for ALL backend development tasks.**

### When to Use goframe-v2 Skill

**MANDATORY** — Invoke the `goframe-v2` skill when:
- Writing or modifying any Go code in `apps/backend/`
- Implementing service layer logic (`internal/service/`)
- Creating new API endpoints or controllers
- Working with database operations (DAO/ORM)
- Implementing error handling, logging, or validation
- Managing configuration, context, or dependency injection
- Any Go backend development task

### Service Layer Requirements

Service layer code (`internal/service/`) MUST follow these patterns:

1. **Interface Definition**: Define service interfaces in `internal/service/{domain}/` with clear method signatures
2. **Implementation Structure**: Implement services with proper dependency injection via `s{ServiceName}` struct
3. **Context Management**: Always pass `ctx context.Context` as the first parameter
4. **Error Handling**: Use GoFrame's `gerror` package for structured error handling
5. **Logging**: Use `g.Log()` with proper context for all logging operations
6. **Configuration Access**: Use `g.Cfg()` for configuration retrieval
7. **Validation**: Use GoFrame's validation tags and `gvalid` package
8. **Transaction Management**: Use `g.DB().Transaction()` for multi-step operations
