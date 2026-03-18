# 项目概述

`Lina`是一个`Go`语言管理后台系统，采用前后端分离架构。

- **前端**: `Vben5 + Vue 3 + Ant Design Vue + TypeScript（pnpm monorepo）`
- **后端**: `GoFrame + MySQL + JWT`
- **参考项目**: `/Users/john/Workspace/gitee/dapppp/ruoyi-plus-vben5`（前端样式和功能交互参考）

## 默认账号

- 用户名称: `admin`
- 登录密码: `admin123`

## 目录结构

```text
apps/                → MonoRepo项目目录
  lina-core/         → GoFrame框架实现的后端源码
    api/             → 请求/响应 DTO（g.Meta 路由定义）
    internal/        → 后端核心代码实现
      cmd/           → 服务启动 & 路由注册
      consts/        → 全局常量定义
      controller/    → HTTP控制器（gf gen ctrl 自动生成骨架）
      dao/           → 数据访问层（gf gen dao 自动生成）
      model/         → 数据模型
        do/          → 数据操作对象（自动生成）
        entity/      → 数据库实体（自动生成）
      service/       → 业务逻辑层
    manifest/        → 交付清单
      config/        → 后端配置文件
      sql/           → DDL + Seed DML（版本 SQL 文件）
        mock-data/   → Mock 演示/测试数据（不随生产部署）
  lina-vben/         → Vben5 前端（pnpm monorepo）
    apps/web-antd/   → 主应用（Ant Design Vue）
    packages/        → 共享库（@core, effects, stores, utils 等）
hack/                → 项目脚本及测试用例文件
  tests/             → E2E 测试（Playwright）
    e2e/             → 测试用例文件
    fixtures/        → 测试 fixtures（auth, config）
    pages/           → 页面对象模型
openspec/            → OpenSpec相关文档
  changes/           → OpenSpec变更记录
```

## 常用命令

### 开发环境

```bash
make dev       # 启动前后端（前端:5666, 后端:8080）
make stop      # 停止所有服务
make status    # 查看服务状态
make test      # 运行完整E2E测试
make init      # 初始化数据库（DDL + Seed 数据）
make mock      # 加载 Mock 演示数据（需先执行 init）
make up        # AI 生成 commit message 并推送
```

### 后端

```bash
cd apps/lina-core
go run main.go          # 运行
make build              # 构建
make dao                # 生成 DAO/DO/Entity（修改 SQL 后）
make ctrl               # 生成控制器骨架（修改 API 定义后）
```

### 前端

```bash
cd apps/lina-vben
pnpm install                   # 安装依赖
pnpm -F @lina/web-antd dev     # 开发模式
pnpm run build                 # 构建
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

# 开发流程规范

本项目采用`SDD`驱动开发，使用`OpenSpec`工具辅助落地。变更记录存放在 `openspec/changes/` 目录下。每个变更包含：`proposal.md`（提案）、`design.md`（设计）、`specs/`（增量规范）、`tasks.md`（任务清单）。

**执行流程**：
1. 通过`/opsx:explore`斜杠指令在给定需求描述的前提下进行探索式对话，分析问题、设计方案、评估风险。
2. 当探索式对话结束，形成清晰的解决方案时，通过`/opsx:propose`斜杠指令将其转化为正式的`OpenSpec`变更提案文档。命令形如`/opsx:propose v1.0.0`，其中`v1.0.0`为当前变更(迭代)的名称。随后会在`openspec/changes`目录下会自动生成一个新的变更文件夹，包含增量规范系列文档(`spec/`)、技术实现方案(`design.md`)、变更提案与思路(`proposal.md`)和实现任务清单(`tasks.md`)。
3. 随后执行`/opsx:apply`开始按照`tasks.md`中的任务清单逐条执行，完成代码实现、测试、文档更新等工作。其中如果涉及前端页面的功能，那么都需要创建`e2e`端到端测试用例，例如数据列表的查询条件、各个按钮的功能实现、列表数据记录的操作按钮等都需要有测试用例，并且在执行过程中自动运行测试用例，确保功能实现的正确性。
4. 执行完成后，可以由人工或`Agent`再次测试产品功能，如果存在问题或者改进点，那么手动调用`/openspec-feedback`技能进行反馈和自动修复、验证，并更新相关`OpenSpec`文档。通过这种方式不断打磨完善当前变更(迭代)的功能，直到没有问题为止。
5. 确认本次变更(迭代)功能已完成没有问题后，则执行`/opsx:archive`斜杠指令将本次变更归档。

**关键规则**：当用户报告问题缺陷/改进建议时（无论中文或英文），如果当前项目存在活跃的`OpenSpec`变更，那么必须调用`openspec-feedback`技能。

# 架构设计规范

## 模块解耦设计

所有前后端模块必须采用解耦设计，业务模块支持按需启用/禁用。设计和实现时须遵循以下原则：

1. **模块可禁用**：每个业务模块（如部门、岗位、字典等）应当是独立的，可以通过配置禁用。禁用某模块后，所有依赖该模块的功能必须自动降级或隐藏，不能出现报错或空白区域。
2. **前端联动隐藏**：当一个模块被禁用时，前端所有涉及该模块的`UI`元素（菜单项、表单字段、表格列、搜索条件、按钮等）必须完全隐藏，而非仅禁用或置灰。例如：禁用"部门"和"岗位"模块后，用户管理页面中不应出现任何部门和岗位相关的筛选条件、表格列或表单字段。
3. **后端松耦合**：后端服务间的依赖应通过接口或可选引用实现，避免硬依赖。当被依赖的模块被禁用时，相关字段返回零值或忽略即可，不应抛出错误。
4. **数据完整性**：模块禁用仅影响功能和展示层，不应删除或破坏已有数据。重新启用模块后，历史数据应能正常恢复使用。

## 接口设计规范

所有前后端`API`必须严格遵循`RESTful`设计规范，`HTTP`方法与操作语义必须一一对应：

| HTTP 方法 | 语义 | 适用场景 |
|-----------|------|---------|
| **GET** | 读取（无副作用） | 列表查询、详情获取、树形数据、导出、下拉选项等所有只读操作 |
| **POST** | 创建资源/执行动作 | 新增记录、文件上传、导入、登录、登出等 |
| **PUT** | 更新资源 | 修改记录、状态变更、重置密码等 |
| **DELETE** | 删除资源 | 单条或批量删除 |

**强制规则**：

1. **查询操作禁止使用POST**：所有查询、列表、搜索、导出、获取详情等读操作必须使用`GET`方法，查询参数通过`URL Query String`传递
2. **创建操作禁止使用GET**：任何会产生副作用（新增数据、上传文件等）的操作禁止使用`GET`方法，必须使用`POST`方法
3. **删除操作必须使用DELETE**：不允许用`POST`或`GET`方法执行删除
4. **更新操作使用PUT**：修改已有资源必须使用`PUT`方法，不允许用`POST`方法
5. **URL 设计使用名词复数或资源名**：如 `/user`、`/dept`、`/dict/type`，避免在 URL 中使用动词（如 `/getUser`、`/deleteUser`）
6. **子资源使用嵌套路径**：如 `/dept/{id}/users`、`/user/{id}/status`

# 代码开发规范

## 后端

### Go代码开发规范
- 所有`Go`后端代码必须使用`goframe-v2`技能开发
- `DAO/DO/Entity`源码文件由`gf gen dao`自动生成，不要手动创建或修改
- `Controller`源码文件由`gf gen ctrl`自动生成骨架，在生成的文件中填写业务逻辑
- **优先使用GoFrame框架提供的组件和方法**：所有`Go`方法调用优先使用`GoFrame`框架已有的方法，避免重复造轮子。例如：
  - 错误处理：使用`GoFrame`的 `gerror` 包进行结构化错误处理
  - 日志记录：使用 `g.Log()` 并传入上下文进行日志记录
  - 配置访问：使用 `g.Cfg()` 获取配置项
  - 数据校验：使用 `GoFrame` 的校验标签和`gvalid`包
  - 遍历目录：使用 `gfile.ScanDirFile`，而非自行实现目录遍历逻辑

### Go代码生成流程
- **API变更**: 修改 `api/{resource}/v1/*.go` → `make ctrl`
- **数据库变更**: 新增或修改 `manifest/sql/{version}.sql`（如 `v0.2.0.sql`）→ `make init`将`sql`文件更新到数据库中 → `make dao`生成或更新`Go`源码文件

### SQL文件管理规范
- **SQL文件命名规范**：每次数据库变更的`SQL`文件以当前迭代版本号命名（如 `v0.1.0.sql`、`v0.2.0.sql`），存放在 `manifest/sql/` 目录下。`init.sql` 仅用于初始建表，后续迭代的表结构变更（`ALTER TABLE`、新增表等）使用版本号命名的`SQL`文件。升级时按版本顺序依次执行即可完成数据库迁移。当前迭代若不涉及数据库变更，则不用生成该迭代的`sql`文件。
- **SQL数据分类管理**：版本`SQL`文件（如 `v0.2.0.sql`）中只允许包含`DDL`（建表/改表）和 `Seed DML`（系统运行所必需的初始化数据，如字典类型、管理员账号等）。演示/测试用的`Mock`数据（如测试用户、演示部门/岗位等）必须放到 `manifest/sql/mock-data/` 目录下的独立`SQL`文件中，文件名以数字前缀控制执行顺序（如 `01_mock_depts.sql`、`02_mock_posts.sql`）。

### 接口层实现要求

接口层代码（`api/`）必须遵循以下模式：

- **接口文件拆解**：在功能模型中，不要将该功能模块的所有的接口都定义到一个`Go`文件中，而应当按照把不同的接口用途拆解到不同的`Go`文件中。例如：用户管理模块中，用户列表查询接口、用户详情接口、用户创建接口等都应该拆解到不同的`Go`文件中，这样可以避免单个`Go`文件过大，导致可读性和维护性变差。

### 服务层实现要求

服务层代码（`internal/service/`）必须遵循以下模式：

- **结构化封装**：使用`Service`作为服务实现的默认结构体名称，当服务层逻辑较复杂时应当解耦拆分为多个结构体来封装业务逻辑
- **上下文管理**：第一个参数始终传入 `ctx context.Context`
- **数据库操作**：
  - **数据交互**：与数据库交互式，必须使用`DO`对象，不使用 `g.Map`来传递`Data`参数
  - **事务管理**：使用 `dao.Xxx.Transaction()`闭包处理多步操作，该方法支持嵌套事务，其中`Xxx`为对应的`Dao`对象名称
  - **跨数据库兼容**：所有数据库操作必须使用跨数据库类型的通用语法，禁止使用特定数据库的内置函数（如`MySQL`的 `FIND_IN_SET`、`GROUP_CONCAT`、`IF()`，`PostgreSQL`的 `ANY(ARRAY[...])`等）。例如对于层级数据（如部门树）的递归查询，应通过应用层迭代查询实现：先通过 `parent_id` 逐层查询收集所有子级`ID`，再使用 `WHERE IN` 进行批量查询，而非依赖数据库特有的递归语法

## 前端

- 路径别名 `#/*` 指向 `./src/*`
- 路由模块放 `src/router/routes/modules/`
- 视图文件放 `src/views/` 对应目录
- API 文件放 `src/api/` 对应目录
- 适配器层 `src/adapter/`：`component`（组件映射）、`form`（表单配置）、`vxe-table`（表格配置）
- 全局组件在 `src/components/global/` 注册（如`GhostButton`用于表格操作列）
- 表格页面使用 `useVbenVxeGrid` + `Page` 组件，操作列用 `ghost-button` + `Popconfirm`
- 前端样式和交互参考`ruoyi-plus-vben5`项目保持一致

## E2E测试要求

- 修复`bug`或新增功能涉及**用户可观察行为变化**时，必须编写或更新对应的`E2E`测试用例
- 修改完成后必须运行相关`E2E`测试并确认通过，再标记任务完成
- 纯内部重构（无`UI`变化）可豁免，但需运行已有测试套件确认无回归
- 使用测试工具（如`Playwright`）在涉及创建文件的场景（如截图）下，应该将创建的文件放置到项目根目录的`temp/`目录下

## UI 设计规范

重要：所有前端 UI 设计和实现必须参考`ruoyi-plus-vben5`项目，保持`UI`的一致性和用户体验的一致性。

在实现任何前端页面或组件时，必须遵循以下规范：

1. **交互设计**: 弹窗（`Modal/Drawer`）、表单、表格、搜索栏等交互模式必须与参考项目保持一致
2. **页面样式**: 布局、间距、字体、颜色等视觉元素参考参考项目的实现
3. **组件使用**: 优先使用与参考项目相同的组件和配置方式，包括：
   - 表单使用 `useVbenForm`，弹窗使用 `useVbenModal`，抽屉使用 `useVbenDrawer`
   - `RadioGroup`单选项使用 `optionType: 'button'` + `buttonStyle: 'solid'`（按钮样式）
   - 文件上传使用 `Upload.Dragger`（拖拽上传样式）
   - 文件下载使用 `requestClient.download` 方法
   - 操作列的"更多"下拉菜单使用 `Dropdown` + `Menu` + `MenuItem`
4. **弹窗规范**: 导入弹窗包含拖拽上传区域、文件类型提示、下载模板链接、覆盖开关；重置密码弹窗包含用户信息展示（`Descriptions`）和密码输入
5. **图标使用**: 使用 `IconifyIcon` 组件（来自 `@vben/icons`），图标名使用`Iconify`格式（如 `ant-design:inbox-outlined`）

开发新页面前，**必须先查看参考项目中对应页面的实现**，确保`UI`和交互保持一致。


