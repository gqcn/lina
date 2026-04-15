# Lina 插件机制设计与开发指南

`apps/lina-plugins/`既是当前插件机制的一期设计文档，也是后续插件开发时的统一参考入口。本文档只描述**仓库中已经落地的真实实现**，同时明确哪些内容仍属于后续规划，避免设计文档、代码实现和开发习惯之间出现偏差。

## 文档定位

本文档同时服务两类读者：

- 插件开发者：需要按照当前约定创建、接入、调试和维护插件。
- 人工 reviewer：需要快速确认某个插件是否符合当前插件框架的设计边界和接入规范。

因此本文档的目标不是“讲概念”，而是回答以下问题：

- 当前插件机制到底支持什么，不支持什么。
- 一个新的源码插件应该放在哪里，哪些文件是必需的。
- 宿主如何发现插件页面、`Slot`、SQL 和后端注册入口。
- `plugin.yaml` 为什么保持最小化，以及哪些字段明确不允许再放进去。
- 开发者在提交插件前应该自查哪些关键点。

## 当前收尾范围

为了让后续人工 review 不再把“当前必须交付的基础项”和“明确后延的工具链/热升级能力”混在一起，这里补充当前收尾口径：

- `apps/lina-plugins/<plugin-id>/` 的标准目录结构、目录职责和 review 要点已经在本文档中落地，因此“目录结构规划”本身不再是缺失项。
- 仓库中的 [plugin-demo-source](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/plugin-demo-source/README.md) 与 [plugin-demo-dynamic](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/plugin-demo-dynamic/README.md) 已经分别覆盖源码插件与动态插件的真实目录形态，后续开发可直接以这两个样例为样板复制和裁剪。
- 当前仍然**不提供额外的自动脚手架脚本和打包脚本**。如果后续评估认为这些脚本会明显增加复杂度、且收益不足，则继续保持“手工维护显式注册关系 + 文档约束 + 样例目录参考”的模式即可。

## 当前范围

当前仓库已经落地的是**第一期：源码插件底座**。插件机制的能力边界如下：

| 能力              | 当前状态 | 说明                                                                                                                                                                                                                                                 |
| ----------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `source`源码插件  | 已实现   | 插件目录位于`apps/lina-plugins/<plugin-id>/`，随宿主一起编译、打包和交付                                                                                                                                                                             |
| `dynamic`动态插件 | 部分实现 | 当前已落地 `wasm` 产物自定义节校验、ABI 版本校验、checksum 与治理元数据同步，并支持上传、安装/卸载 SQL 执行、前端静态资源内存 bundle 托管，以及基于托管资源的 `iframe` / 新标签页 / 宿主内嵌挂载三种页面接入；热装载、真正的运行时执行与回滚仍未交付 |
| 插件管理页        | 已实现   | 支持源码插件同步、启用、禁用与治理联动                                                                                                                                                                                                               |
| 后端扩展点        | 已实现   | 通过`pluginhost`发布的回调式扩展点接入                                                                                                                                                                                                               |
| 前端页面接入      | 已实现   | 扫描`frontend/pages/**/*.vue`并挂到宿主运行时页                                                                                                                                                                                                      |
| 前端`Slot`接入    | 已实现   | 扫描`frontend/slots/**/*.vue`并挂到宿主公开插槽                                                                                                                                                                                                      |
| 插件安装 SQL      | 已实现   | 通过`manifest/sql/*.sql`目录约定发现                                                                                                                                                                                                                 |
| 插件卸载 SQL      | 已实现   | 通过`manifest/sql/uninstall/*.sql`目录约定发现                                                                                                                                                                                                       |
| 样例插件样板      | 已提供   | 提供 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录，帮助开发者按当前真实契约复制和裁剪新插件                                                                                                                                        |
| 脚手架脚本        | 未提供   | 当前不再提供`hack/plugin`下的自动创建脚本；如果这些脚本会增加复杂度且收益很低，则继续保持不引入                                                                                                                                                      |

补充说明：

- 运维与 review 说明已整理到 [OPERATIONS.md](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/OPERATIONS.md)。
- 当前仓库提供两个 review 样例：
  - `plugin-demo-source`：源码插件样例
  - `plugin-demo-dynamic`：动态插件样例
  - 其中 `plugin-demo-source` 当前只保留左侧菜单页示例；动态页面挂载、独立静态页等对比能力由 `plugin-demo-dynamic` 演示

## 设计原则

当前插件机制遵循以下原则：

### 约定优于配置

- 前端页面位置、前端`Slot`位置、安装 SQL 和卸载 SQL 都通过固定目录约定发现。
- `plugin.yaml` 只额外承担插件菜单元数据声明；其余可推导信息不再重复配置。

### 单一真相源

- 菜单、按钮权限和父子关系以 `plugin.yaml` 的 `menus` 元数据为单一真相源。
- 后端扩展能力以插件代码注册为单一真相源。
- 前端页面和`Slot`以真实源码文件为单一真相源。

### 显式接线

- 当前源码插件的后端接线方式不是脚本生成，也不是隐式自动装配。
- 开发者需要显式维护`apps/lina-plugins/lina-plugins.go`，让宿主编译期导入插件后端包。
- 这样做的目的是让接线关系清晰、可 grep、可 review、可追踪。
- 在当前阶段，这种“手工接线 + 文档约束”的模式优先级高于额外引入脚手架或打包脚本。

### 设计与实现一致

- 文档中不再保留已经移除的元数据模型和自动化脚本描述。
- 对未来能力的描述会明确标注为“规划中”，不能和“已实现”混写。

## 插件类型

当前插件一级类型只保留两类：

| 类型      | 含义                                             | 当前状态 |
| --------- | ------------------------------------------------ | -------- |
| `source`  | 源码插件，目录在`apps/lina-plugins/<plugin-id>/` | 已实现   |
| `dynamic` | 动态插件，面向后续热安装与热升级                 | 部分实现 |

重要说明：

- 当前动态插件已经补齐了上传、安装/卸载 SQL、静态资源托管以及**宿主可控的 declarative backend execution** 基线，但它仍然不是“任意插件 Go 代码直接热执行”的能力模型。
- 历史上把`wasm`当一级类型的设计已经收敛掉了。当前治理视角只区分`source`和`dynamic`。
- 如果当前要开发新插件，默认应按照`source`源码插件方式开发。

## 当前 `runtime wasm` 产物约定

当前仓库已经交付“上传 `wasm` -> 安装/启用 -> 前端托管 -> declarative backend Hook / 资源执行”的二期运行时链路，并把**产物契约与宿主校验规则**真正落到了代码里，便于人工 review。

为了让 reviewer 能看到一个“不是抽象 JSON，而是实际资源文件”的动态样例，仓库提供了独立的 [plugin-demo-dynamic](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/plugin-demo-dynamic/README.md) 插件目录。该目录现在直接以插件源码树作为作者侧真相源：`plugin_embed.go` 通过 `go:embed` 统一声明 `plugin.yaml`、`frontend/`、`manifest/` 等静态资源，`hack/build-wasm` 再把这些资源转换为宿主可治理的 Wasm 自定义节快照；标准仓库构建流程会把最终产物统一输出到仓库根 `temp/output/`，自动化测试会验证生成产物与明文源码保持一致。

对于当前已经落地的宿主服务命名，插件作者可以直接按下面的心智模型理解：

- `runtime`：日志、运行时状态、宿主时间与节点信息。
- `storage`：按逻辑路径或路径前缀授权的对象/文件读写。
- `network`：按 URL pattern 授权的出站 HTTP 请求。
- `data`：按宿主确认授权的数据表执行结构化数据访问，而不是直接连接数据库。

其中 `data` service 的能力命名已经固定为“两类 capability + 六个 method”：

- `host:data:read`：表示数据读取类能力，对应 `list` 和 `get`
- `host:data:mutate`：表示数据变更类能力，对应 `create`、`update`、`delete` 和 `transaction`
- `list`：按过滤条件分页查询多条记录
- `get`：按主键或唯一键读取单条记录

当前不再对动态插件暴露 `host:data:query`、raw SQL、通用 SQL 执行接口或数据库直连能力。插件若需要访问宿主数据，必须在 `plugin.yaml` 中通过 `hostServices` 声明 `service: data`、所需 `methods` 以及 `resources.tables`，再由宿主在安装/启用时确认最终授权。

作者侧清单建议补充两条约束：

- `plugin.yaml` 中本迭代新增的 `hostServices`、`resources.paths`、`resources.tables`、URL pattern 等字段，应在样例里提供就地注释说明，避免 reviewers 需要跳回文档对照理解。
- guest 后端的控制器应保持轻量，复杂业务逻辑统一放在 `backend/internal/service/<component>/` 中维护，控制器只负责桥接请求上下文与响应装配。

对于 guest 侧的实际编码方式，当前推荐优先使用 `lina-core/pkg/plugindb` 提供的受限 ORM 风格 facade，而不是继续直接调用底层 `pluginbridge.Data()` helper。推荐心智模型如下：

```go
db := plugindb.Open()

records, total, err := db.Table("sys_plugin_node_state").
    Fields("id", "nodeKey", "currentState").
    WhereEq("pluginId", pluginID).
    WhereLike("nodeKey", "demo-").
    WhereIn("currentState", []string{"pending", "running"}).
    OrderDesc("id").
    Page(1, 10).
    All()
```

约束如下：

- `plugindb` 只提供单表、受治理、结构化的数据访问体验；
- 首期开放的高层过滤操作主要包括 `WhereEq`、`WhereIn`、`WhereLike`，排序包括 `OrderAsc`、`OrderDesc`；
- `Update`、`Delete` 默认要求通过 `WhereKey(...)` 指定目标记录；
- `Transaction(...)` 当前只支持单表结构化 mutation；
- 底层仍然通过宿主 `data` hostService 执行，不会向插件暴露完整 `gdb.DB`、`gdb.Model` 或宿主 DAO。

当前约定如下：

- `apps/lina-plugins/<plugin-id>/` 仅作为动态样例插件的明文源码与构建输入目录，不再作为宿主运行时发现入口。
- 动态插件作者推荐在根目录提供 `plugin_embed.go`，通过 `go:embed plugin.yaml frontend manifest` 统一声明需要随 `.wasm` 交付的静态资源。
- 宿主当前通过配置项 `plugin.dynamic.storagePath` 发现和管理 runtime `wasm` 插件，默认值为 `temp/output`。
- 当 `storagePath` 使用相对路径时，宿主会以仓库根目录作为解析基准，保证上传、手工拷贝和后台同步识别走同一目录。
- 宿主只扫描 `storagePath` 根目录下的 `*.wasm` 文件，不对外层目录层级做额外约定。
- 动态插件上传后，宿主会以 `<storagePath>/<plugin-id>.wasm` 的规范文件名落盘；若运维手工拷贝 `.wasm` 到该目录，则可通过管理页“同步插件”识别。
- 仓库标准构建入口会把 `apps/lina-plugins/<plugin-id>/` 的动态样例产物统一输出到 `temp/output/<plugin-id>.wasm`；若要被宿主识别，仍需上传或复制到 `plugin.dynamic.storagePath`。
- `plugin.dynamic.storagePath` 当前是节点本地文件系统目录语义；宿主不会把上传到某一节点的 `.wasm` 自动复制到其他节点。
- 这类约束不仅适用于动态插件 `wasm`；宿主其他写入本地磁盘的上传资源同样不会自动跨节点分发，例如 `upload.path` 下的通用上传文件与其他本地静态资源。
- 因此，多节点部署时必须为 `plugin.dynamic.storagePath`、`upload.path` 及其他需跨节点访问的资源目录配置共享存储，或在宿主之外完成可靠的文件分发；否则从节点上传的资源可能只在该节点可见，主节点安装、其他节点收敛或用户访问资源都可能失败。
- 宿主当前会读取 `storagePath/*.wasm` 中两个必选自定义节：
  - `lina.plugin.manifest`
  - `lina.plugin.dynamic`
- 宿主当前还支持一个可选前端资源自定义节：
  - `lina.plugin.frontend.assets`
- 宿主当前还支持两个可选 SQL 自定义节：
  - `lina.plugin.install.sql`
  - `lina.plugin.uninstall.sql`
- 宿主当前还支持两个可选后端契约自定义节：
  - `lina.plugin.backend.hooks`
  - `lina.plugin.backend.resources`
- 这些自定义节当前统一使用 JSON 编码，便于 reviewer 直接理解其字段语义。
- 构建器当前会优先消费动态插件的 `go:embed` 资源声明，再回退到旧的目录扫描方式；宿主本身不直接读取 guest `embed.FS`，仍只消费构建后产物里的静态快照。

`lina.plugin.manifest` 当前要求至少包含：

```json
{
  "id": "plugin-demo-dynamic",
  "name": "动态插件示例",
  "version": "v0.1.0",
  "type": "dynamic"
}
```

`lina.plugin.dynamic` 当前要求至少包含：

```json
{
  "runtimeKind": "wasm",
  "abiVersion": "v1",
  "frontendAssetCount": 0,
  "sqlAssetCount": 0
}
```

如果动态插件需要携带迁移 SQL，可以额外嵌入如下 JSON 数组：

```json
[
  {
    "key": "001-plugin-demo-dynamic.sql",
    "content": "CREATE TABLE plugin_demo_record (...);"
  }
]
```

约束如下：

- `key` 必须符合宿主同款 SQL 命名规范：`{序号}-{当前迭代名称}.sql`
- `key` 不能包含目录分隔符
- `content` 不能为空
- 宿主当前会保持数组顺序，并按顺序执行 install / uninstall SQL 步骤

如果动态插件需要声明宿主可执行的后端 Hook，可以额外嵌入如下 JSON 数组：

```json
[
  {
    "event": "auth.login.succeeded",
    "action": "insert",
    "mode": "blocking",
    "timeoutMs": 1000,
    "table": "plugin_runtime_login_log",
    "fields": {
      "user_name": "event.userName",
      "created_at": "now"
    }
  }
]
```

当前后端 Hook 契约约束如下：

- `event` 必须是宿主已发布的 Hook 插槽。
- `mode` 当前仅允许宿主已发布的执行模式；Hook 支持 `blocking` 与 `async`。
- `timeoutMs` 可选；若不声明，宿主使用默认超时。
- `action` 当前仅支持宿主内建动作：
  - `insert`：向插件自有表插入一行审计/事件数据
  - `sleep`：等待指定 `sleepMs`，主要用于验证 timeout / isolation 行为
  - `error`：直接返回一个声明式错误，主要用于验证 failure isolation 行为
- `insert` 动作要求 `table` 与 `fields` 合法。
- `sleep` 动作要求 `sleepMs > 0`。
- `error` 动作要求 `errorMessage` 非空。

如果动态插件需要声明宿主可查询的后端资源，可以额外嵌入如下 JSON 数组：

```json
[
  {
    "key": "records",
    "type": "table-list",
    "table": "plugin_runtime_records",
    "fields": [
      { "name": "id", "column": "id" },
      { "name": "title", "column": "title" }
    ],
    "orderBy": {
      "column": "id",
      "direction": "asc"
    },
    "dataScope": {
      "userColumn": "owner_user_id",
      "deptColumn": "owner_dept_id"
    }
  }
]
```

当前后端资源契约约束如下：

- `type` 当前仅支持 `table-list`。
- `table`、`fields`、`filters`、`orderBy` 中涉及的表名和列名都必须通过宿主标识校验。
- `dataScope` 可选；一旦声明，宿主会将当前登录用户的角色数据权限自动应用到该资源查询。
- `dataScope.userColumn` 对应“仅本人”过滤。
- `dataScope.deptColumn` 对应“本部门”过滤。
- 若当前用户无可用角色数据权限，或插件声明的数据权限列不足以支撑当前范围，宿主会按最小权限原则返回空结果。

宿主当前会执行以下校验：

- 宿主识别到的 `.wasm` 文件必须存在于 `plugin.dynamic.storagePath`。
- `wasm` 文件头和版本必须合法。
- 必须包含 `lina.plugin.manifest` 和 `lina.plugin.dynamic` 两个自定义节。
- 若声明了 `lina.plugin.frontend.assets`，宿主会校验每个前端资源的 `path`、`contentBase64`，并拒绝路径越界。
- 嵌入清单中的 `id/name/version/type` 必须完整且合法，且 `type` 当前只能为 `dynamic`。
- `runtimeKind` 当前只能是 `wasm`。
- `abiVersion` 当前只能是 `v1`。
- 若存在 `lina.plugin.install.sql` / `lina.plugin.uninstall.sql`，宿主会校验每个 SQL 资源键和值是否合法。
- 若存在 `lina.plugin.backend.hooks`，宿主会校验每个 Hook 的 `event/action/mode/timeoutMs` 及其动作专属字段是否合法。
- 若存在 `lina.plugin.backend.resources`，宿主会校验资源键、字段、过滤器、排序和 `dataScope` 列声明是否合法。

当前已经落地到治理元数据中的内容包括：

- `sys_plugin.checksum` 会记录运行时产物的 `sha256`。
- `sys_plugin_release.runtime_kind` 会记录 `wasm`。
- `manifest_snapshot` 会记录 `runtimeKind`、`runtimeAbiVersion`、`runtimeFrontendAssetCount`、`runtimeSqlAssetCount`。
- `sys_plugin_resource_ref` 会新增一条抽象的 `runtime_wasm` 资源摘要，方便 reviewer 确认宿主确实识别到了运行时产物。
- 若运行时产物声明了前端静态资源，`sys_plugin_resource_ref` 还会新增 `runtime_frontend` 摘要，便于 reviewer 确认宿主识别到了多少个可托管资源。
- `sys_plugin_migration` 在动态插件安装/卸载时，会优先针对嵌入在 `wasm` 中的 SQL 资源记录执行结果；只有当未声明嵌入 SQL 时，才会回退到目录约定 SQL。
- 动态插件 Hook 在宿主执行时，会统一经过 timeout、error 与 panic isolation 包装，不允许单个动态插件阻断宿主主流程。
- runtime 资源查询会复用宿主当前登录用户的角色数据权限上下文，而不是绕开 Lina 现有治理模型。

### 当前已提供的前端静态资源托管基线

如果动态插件在 `lina.plugin.frontend.assets` 中嵌入了前端静态资源，宿主当前会直接以 `plugin.dynamic.storagePath` 中的 `.wasm` 作为单一真相源，在内存中构建只读资源视图并对外提供稳定访问地址。服务重启后，宿主会在启动时预热已安装且已启用的动态插件资源；若某个 bundle 未预热成功，请求链路仍会按需重新从 `.wasm` 懒加载。

资源对外公开路径固定为：

```text
/plugin-assets/<plugin-id>/<version>/<relative-path>
```

例如：

```text
/plugin-assets/plugin-demo-dynamic/v0.1.0/standalone.html
```

当前托管边界如下：

- 只有 `dynamic` 插件才允许走这条公开资源路径。
- 请求中的 `<version>` 必须与宿主当前识别到的插件版本一致。
- 插件必须处于“已安装 + 已启用”状态，否则宿主会返回不可访问。
- 当请求路径为空时，宿主默认回退到 `index.html`。
- 当插件在 manifest `menus` 元数据中将 `path` 写成 `/plugin-assets/<plugin-id>/<version>/...` 这类托管地址时，宿主当前已经支持两种菜单驱动接入模式：
  - `is_frame = 0`：自动转换为 `iframe` 路由，托管页面在宿主主内容区内嵌打开。
  - `is_frame = 1`：自动转换为新标签页路由，点击菜单后直接打开托管地址。
- 当插件菜单同时满足以下条件时，宿主会走第三种“宿主内嵌挂载”模式：
  - `component = system/plugin/dynamic-page`
  - `path = /plugin-assets/<plugin-id>/<version>/<entry-file>`
  - `query_param` 包含 `{"pluginAccessMode":"embedded-mount"}`
- 进入该模式后，宿主不会把菜单当成 `iframe` 或新标签页，而是：
  - 为菜单生成一个宿主内部动态路由路径
  - 将原始托管资源地址透传为 `embeddedSrc` 查询参数
  - 由 `system/plugin/dynamic-page` 壳在宿主内容容器内动态导入该 ESM 入口
- 动态插件在启用前还会额外校验这些菜单引用的托管资源是否真实存在；若菜单引用了缺失资源，或宿主内嵌挂载入口不是 `.js/.mjs` ESM 文件，则启用会被拒绝。

### 当前已提供的宿主内嵌挂载契约

当前动态插件的宿主内嵌挂载不是“任意 HTML 页面直接塞进 DOM”，而是一个**最小 ESM 契约**。宿主当前要求被内嵌的托管入口至少导出：

```ts
export function mount(context) {
  // render into context.container
}
```

当前宿主会传入的 `context` 至少包括：

- `container`: 宿主为当前插件页面准备好的挂载 DOM 容器
- `assetURL`: 当前 ESM 入口的完整访问地址
- `baseURL`: 当前 ESM 入口所在目录的基础 URL，便于插件自行拼接其他静态资源
- `routePath`: 当前宿主动态路由路径
- `title`: 当前菜单标题
- `query`: 宿主路由层合并后的查询参数

当前宿主还支持两个可选生命周期能力：

```ts
export function unmount(context) {}
export function update(context) {}
```

或者由 `mount(context)` 返回一个对象：

```ts
export function mount(context) {
  return {
    unmount(nextContext) {},
    update(nextContext) {},
  };
}
```

当前实现边界如下：

- 宿主只负责“准备容器 + 动态导入 ESM + 调用 mount/unmount/update”。
- 宿主不会替插件注入额外 SDK、状态管理、路由桥接或主题桥接。
- 若插件需要更复杂的宿主能力，当前仍应优先退回 `iframe` 或新标签页模式。
- 宿主当前会在页面内显式展示挂载失败结果，便于人工 review 判断到底是菜单接入问题还是插件入口自身实现问题。

### 当前已提供的上传入口

宿主当前已经提供动态插件包上传 API：

```text
POST /api/v1/plugins/dynamic/package
Content-Type: multipart/form-data
```

表单字段如下：

- `file`: 必填，`.wasm` 文件
- `overwriteSupport`: 可选，`1` 表示允许覆盖**尚未安装**的同 ID 动态插件文件

当前上传接口的行为边界如下：

- 宿主会先解析上传包中的嵌入清单和运行时元数据。
- 宿主会将产物直接写入 `plugin.dynamic.storagePath/<plugin-id>.wasm`，不再在 `apps/lina-plugins/` 下额外生成 `plugin.yaml` 或运行时工作目录。
- 若 `wasm` 中声明了 `lina.plugin.frontend.assets`，宿主会在运行时直接从 `.wasm` 解析这些资源，并按需刷新内存缓存。
- 宿主会立即同步 `sys_plugin` / `sys_plugin_release` / `sys_plugin_resource_ref` 等治理元数据。
- 上传完成后，插件默认仍是“未安装、未启用”状态。
- 当前**不允许**通过上传直接覆盖一个已经安装的动态插件；升级/回滚的正式 release 切换能力还没有交付。
- 除上传外，运维也可以手工把 `.wasm` 文件复制到 `plugin.dynamic.storagePath`，然后在管理页执行同步识别。
- 若当前部署为多节点且 `plugin.dynamic.storagePath` 没有使用共享存储，上传请求落到哪一个节点，`.wasm` 文件就只会先写到哪一个节点；宿主当前不会自动把该文件同步到主节点或其他从节点。
- 因此，多节点环境下应当将上传入口、手工拷贝流程与共享存储或外部分发流程一并设计好，再执行安装、启用、升级等生命周期动作；否则数据库中的治理记录可能已经存在，但主节点或其他节点仍然拿不到对应产物。

当前仓库同时提供通用构建入口，供样例动态插件生成宿主可扫描的 wasm 产物：

```bash
make wasm
make wasm p=plugin-demo-dynamic
```

其中：

- `make wasm` 会遍历 `apps/lina-plugins/` 下所有 `type: dynamic` 的插件目录并生成 `temp/output/<plugin-id>.wasm`
- `make wasm p=<plugin-id>` 只构建指定动态插件
- `make wasm` 当前直接通过根级 `hack/build-wasm/` 独立 Go 工具生成产物；该工具有自己的 `go.mod`，不依赖 `apps/lina-core`
- 生成产物不会提交到 Git，`.gitignore` 已忽略仓库根 `temp/`
- 根级 `make dev` / `make build` 会自动复用同一个通用构建入口，避免因为仓库中不提交 wasm 而导致宿主扫描失败

当前仍然**没有**落地的能力包括：

- 在宿主进程内热装载动态插件逻辑。
- 运行时升级、回滚和多节点代际切换。

## 源码插件生命周期

源码插件和动态插件的生命周期语义并不相同。源码插件当前遵循下表：

| 动作 | 源码插件行为                                             |
| ---- | -------------------------------------------------------- |
| 发现 | 宿主扫描`apps/lina-plugins/*/plugin.yaml`识别插件        |
| 同步 | 宿主同步`sys_plugin`记录，保持插件列表和实际目录一致     |
| 安装 | 不提供。源码插件视为随宿主编译即已集成                   |
| 卸载 | 不提供。移除源码插件需要修改源码目录和注册关系后重新构建 |
| 启用 | 已支持。启用后路由、菜单、页面和`Slot`恢复生效           |
| 禁用 | 已支持。禁用后路由、菜单、页面和`Slot`隐藏或拒绝访问     |

这意味着：

- 插件管理页中，源码插件不应出现“安装”“卸载”按钮。
- 新增一个源码插件后，如果目录、清单和注册关系都正确，宿主同步后会把它视为已集成插件。
- 禁用插件不会删除已有业务数据；重新启用后，应能恢复原有治理关系。

## 目录结构

当前源码插件统一放在`apps/lina-plugins/<plugin-id>/`下。推荐目录如下：

```text
apps/lina-plugins/
  README.md
  lina-plugins.go
  <plugin-id>/
    go.mod
    plugin.yaml
    README.md
    backend/
      plugin.go
      api/
      internal/
        controller/
      service/
    frontend/
      pages/
        *.vue
      slots/
        <slot-key>/
          *.vue
    manifest/
      sql/
        001-<iteration-name>.sql
        uninstall/
          001-<iteration-name>.sql
```

各目录职责如下：

| 路径                                       | 作用                       | 是否必需          |
| ------------------------------------------ | -------------------------- | ----------------- |
| `apps/lina-plugins/lina-plugins.go`        | 宿主源码插件后端导入注册表 | 是                |
| `<plugin-id>/go.mod`                       | 插件独立 Go 模块声明       | `source`插件必需  |
| `<plugin-id>/plugin.yaml`                  | 插件最小元数据清单         | 是                |
| `<plugin-id>/README.md`                    | 插件自身说明文档           | 强烈建议          |
| `<plugin-id>/backend/plugin.go`            | 插件后端注册入口           | `source`插件必需  |
| `<plugin-id>/backend/api/`                 | 插件 API 定义              | 按需              |
| `<plugin-id>/backend/internal/controller/` | 插件控制器实现             | 按需              |
| `<plugin-id>/backend/service/`             | 插件服务层实现             | 按需              |
| `<plugin-id>/frontend/pages/`              | 插件页面源码目录           | 有页面时必需      |
| `<plugin-id>/frontend/slots/`              | 插件`Slot`源码目录         | 有`Slot`时必需    |
| `<plugin-id>/manifest/sql/`                | 插件安装 SQL 目录          | 有安装 SQL 时必需 |
| `<plugin-id>/manifest/sql/uninstall/`      | 插件卸载 SQL 目录          | 有卸载 SQL 时必需 |

### 样例插件作为样板

仓库当前不再额外维护 `hack/plugin-template/` 目录，而是直接使用已经接入宿主的真实样例插件作为开发样板：

- 源码插件样板：[plugin-demo-source](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/plugin-demo-source/README.md)
- 动态插件样板：[plugin-demo-dynamic](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/plugin-demo-dynamic/README.md)

建议使用方式：

1. 根据目标类型复制 `plugin-demo-source` 或 `plugin-demo-dynamic` 的整体目录
2. 统一替换插件 ID、菜单 key、权限码、路由路径和文案
3. 根据实际业务重命名内部模块目录与文件
4. 若新增源码插件，在 [apps/lina-plugins/lina-plugins.go](/Users/john/Workspace/github/gqcn/lina/apps/lina-plugins/lina-plugins.go) 中追加匿名导入
5. 删除不需要的样例页面、样例 API 和样例 SQL，占位内容不必长期保留

## 元数据底座

为了让后续人工 review 不必只依赖日志，宿主当前会把插件治理元数据同步到以下表中：

| 表名                      | 当前用途                                         |
| ------------------------- | ------------------------------------------------ |
| `sys_plugin`              | 插件注册表，记录插件基础状态                     |
| `sys_plugin_release`      | 记录插件版本、清单基础信息和资源数量摘要快照     |
| `sys_plugin_migration`    | 记录安装/卸载迁移的执行结果与抽象执行键          |
| `sys_plugin_resource_ref` | 记录宿主发现到的抽象资源类型、稳定标识与摘要说明 |
| `sys_plugin_node_state`   | 记录当前节点对插件状态的观测结果                 |

这些表的目标不是把二三期能力一次性做完，而是先把后续 runtime 生命周期需要的宿主元数据底座稳定下来。

同时需要明确当前持久化边界：

- 宿主会按目录约定扫描 SQL、页面和 `Slot`，但这些具体文件路径只用于校验与执行，不写入插件治理表。
- `manifest_snapshot` 只保存基础清单字段、是否声明清单以及各类资源数量摘要。
- `sys_plugin_resource_ref` 只保存抽象资源键、owner 标识和 summary remark，不保存具体前端文件路径或 SQL 文件路径。
- `sys_plugin_migration` 只保存类似 `install-step-001` 的抽象迁移执行键，不保存具体 SQL 相对路径。

当前插件管理页已经基于这些表补齐了以下治理摘要字段，便于人工 review：

| 字段             | 说明                                                       |
| ---------------- | ---------------------------------------------------------- |
| `releaseVersion` | 宿主当前视角下的生效版本号                                 |
| `lifecycleState` | 生命周期状态键，如 `source_enabled`、`runtime_installed`   |
| `nodeState`      | 当前节点观测状态，如 `enabled`、`installed`、`uninstalled` |
| `resourceCount`  | 当前生效版本登记的资源引用数量                             |
| `migrationState` | 最近一次迁移结果，如 `none`、`succeeded`、`failed`         |

## `plugin.yaml`

### 设计目标

当前`plugin.yaml`故意保持最小化。它的职责只有两类：

- 声明“这个目录是一个插件”。
- 提供插件在治理侧展示和校验所需的基础身份信息。
- 在需要宿主接入菜单时，提供精简的`menus`声明。

它**不再负责**：

- 声明页面入口。
- 声明前端`Slot`。
- 声明 SQL 文件列表。
- 声明宿主兼容矩阵、脚本入口、打包入口。

### 推荐示例

```yaml
id: plugin-demo-source
name: 源码插件示例
version: v0.1.0
type: source
description: 提供左侧菜单页面与公开/受保护路由示例的源码插件
menus:
  - key: plugin:plugin-demo-source:sidebar-entry
    name: 源码插件示例
    path: plugin-demo-source-sidebar-entry
    component: system/plugin/dynamic-page
    perms: plugin-demo-source:example:view
    icon: ant-design:appstore-outlined
    type: M
    sort: -1
    remark: 源码插件示例左侧菜单
author: lina-team
homepage: https://example.com/lina/plugins/plugin-demo-source
license: Apache-2.0
```

### 字段说明

| 字段          | 是否必填 | 说明                                                                |
| ------------- | -------- | ------------------------------------------------------------------- |
| `id`          | 是       | 插件稳定标识，必须使用`kebab-case`，且在宿主范围内唯一              |
| `name`        | 是       | 插件显示名称                                                        |
| `version`     | 是       | 插件版本号，必须使用`semver`格式；本文档示例统一使用带`v`前缀的写法 |
| `type`        | 是       | 当前仅允许`source`或`dynamic`                                       |
| `description` | 否       | 插件简要描述，建议明确功能边界                                      |
| `menus`       | 否       | 插件菜单元数据；源码插件同步和动态插件安装/卸载时由宿主据此治理菜单 |
| `author`      | 否       | 插件作者或团队标识                                                  |
| `homepage`    | 否       | 插件主页或项目地址                                                  |
| `license`     | 否       | 插件许可信息                                                        |

### 宿主校验规则

宿主当前会对`plugin.yaml`做以下校验：

| 校验项             | 规则                                                                        |
| ------------------ | --------------------------------------------------------------------------- |
| `id` 非空          | 缺失则判定清单非法                                                          |
| `id` 格式          | 必须匹配`^[a-z0-9]+(?:-[a-z0-9]+)*$`                                        |
| `id` 唯一性        | 不允许两个插件目录使用同一个`id`                                            |
| `name` 非空        | 缺失则判定清单非法                                                          |
| `version` 非空     | 缺失则判定清单非法                                                          |
| `version` 格式     | 必须满足`semver`格式，例如`v0.1.0`；宿主当前同时兼容不带`v`前缀的写法       |
| `type` 合法性      | 仅允许`source`或`dynamic`                                                   |
| `menus` 合法性     | 若声明菜单，则 `key` 必须使用当前插件前缀，且 `parent_key` 不得引用其他插件 |
| `source`目录完整性 | `source`插件必须存在`go.mod`和`backend/plugin.go`                           |

### 明确不再允许的字段

以下字段已经被当前设计明确淘汰，不应再写入`plugin.yaml`：

- `schemaVersion`
- `compatibility`
- `entry`
- `capabilities`
- `resources`
- `metadata`

这些字段被移除的原因是它们会把以下信息重复建模；其中动态插件作者侧的宿主能力申请现在统一只通过 `hostServices` 维护，不再保留单独的 `capabilities` 顶层输入：

- SQL 文件路径，本来就可以从固定目录推导。
- 前端页面和`Slot`文件，本来就可以从真实源码目录推导。
- 路由和扩展点接入，本来就应该以插件代码注册为真相源。

## 后端接入

### 总体模型

源码插件的后端接入是“插件目录内实现 + `pluginhost`注册 + 宿主显式导入”三段式模型：

1. 插件在`backend/plugin.go`里创建并注册`SourcePlugin`。
2. 插件通过`pluginhost`向宿主注册路由、Hook、过滤器等回调。
3. 宿主通过`apps/lina-plugins/lina-plugins.go`匿名导入插件后端包，让其`init()`逻辑参与宿主编译产物。

### 宿主导入注册表

当前导入注册表文件是：

```go
package linaplugins

import (
	_ "lina-plugin-demo-source/backend"
)
```

新增插件时，开发者需要手工追加匿名导入，例如：

```go
package linaplugins

import (
	_ "lina-plugin-demo-source/backend"
	_ "lina-plugin-foo/backend"
)
```

这是当前源码插件后端接入的**唯一显式接线点**。

### `backend/plugin.go` 最小示例

```go
package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
)

const pluginID = "plugin-demo-source"

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	middlewares := registrar.Middlewares()

	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)
	})
	return nil
}
```

### 当前已发布的后端扩展点

宿主当前已经正式发布的后端扩展点如下：

| Go 常量                              | Canonical 值              | 类型      | 支持模式            |
| ------------------------------------ | ------------------------- | --------- | ------------------- |
| `ExtensionPointAuthLoginSucceeded`   | `auth.login.succeeded`    | 事件 Hook | `blocking`、`async` |
| `ExtensionPointAuthLoginFailed`      | `auth.login.failed`       | 事件 Hook | `blocking`、`async` |
| `ExtensionPointAuthLogoutSucceeded`  | `auth.logout.succeeded`   | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginInstalled`      | `plugin.installed`        | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginEnabled`        | `plugin.enabled`          | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginDisabled`       | `plugin.disabled`         | 事件 Hook | `blocking`、`async` |
| `ExtensionPointPluginUninstalled`    | `plugin.uninstalled`      | 事件 Hook | `blocking`、`async` |
| `ExtensionPointSystemStarted`        | `system.started`          | 事件 Hook | `blocking`、`async` |
| `ExtensionPointHTTPRouteRegister`    | `http.route.register`     | 注册点    | `blocking`          |
| `ExtensionPointHTTPRequestAfterAuth` | `http.request.after-auth` | 注册点    | `blocking`          |
| `ExtensionPointCronRegister`         | `cron.register`           | 注册点    | `blocking`          |
| `ExtensionPointMenuFilter`           | `menu.filter`             | 注册点    | `blocking`          |
| `ExtensionPointPermissionFilter`     | `permission.filter`       | 注册点    | `blocking`          |

开发约束：

- 事件 Hook 可以使用`blocking`或`async`。
- 注册式扩展点当前只允许`blocking`。
- 如果为扩展点声明了不支持的执行模式，宿主会在注册阶段拒绝。

### `RouteRegistrar` 能力

插件路由注册当前通过`RouteRegistrar`完成。它提供两类能力：

| 能力                | 说明                               |
| ------------------- | ---------------------------------- |
| `Group(prefix, fn)` | 在宿主插件路由根分组下创建路由分组 |
| `Middlewares()`     | 获取宿主已发布的中间件目录         |

当前可供插件组合的宿主中间件包括：

- `NeverDoneCtx()`
- `HandlerResponse()`
- `CORS()`
- `Ctx()`
- `Auth()`
- `OperLog()`

重要语义：

- 插件路由本身受插件启停状态保护。插件被禁用后，宿主会在路由入口处直接拒绝访问。
- 宿主不会为插件自动追加固定前缀。插件自己决定是否挂到`/api/v1`或其他前缀下。
- 同一个插件可以在一次注册中拆分多个分组，分别挂载匿名和鉴权路由。

### `CronRegistrar` 能力

如果插件需要注册定时任务，可以使用`RegisterCron`和`CronRegistrar`：

| 能力                               | 说明                             |
| ---------------------------------- | -------------------------------- |
| `Add(ctx, pattern, name, handler)` | 注册一个受插件启停保护的定时任务 |
| `IsPrimaryNode()`                  | 返回当前节点是否为主节点         |

建议：

- 如果任务只应该在主节点执行，插件应自行在回调内通过`IsPrimaryNode()`做判断。
- 定时任务的业务逻辑应放在插件自己的服务层，不要把大段业务逻辑堆在注册回调里。

### 插件后端资源声明

当前源码插件仍支持通过`RegisterResource`声明后端资源，以便复用宿主通用资源查询接口：

- 资源声明在插件代码中完成，而不是在`plugin.yaml`中配置。
- 资源查询统一走宿主的`GET /plugins/{id}/resources/{resource}`契约。

如果插件不需要暴露这类统一资源接口，可以完全不注册。

## 前端页面接入

### 目录约定

插件页面统一放在：

```text
frontend/pages/**/*.vue
```

宿主构建时会扫描这些页面源码，并将其挂载到插件运行时页面容器中。

### `pluginPageMeta`

页面文件可以通过导出`pluginPageMeta`提供显式元数据，例如：

```vue
<script lang="ts">
export const pluginPageMeta = {
  routePath: "plugin-demo-source-sidebar-entry",
  title: "源码插件示例",
};
</script>
```

当前支持的页面元数据字段如下：

| 字段        | 是否必填 | 说明                             |
| ----------- | -------- | -------------------------------- |
| `pluginId`  | 否       | 不传时默认从文件路径推导         |
| `routePath` | 否       | 不传时宿主会根据文件路径自动推导 |
| `title`     | 否       | 不传时默认使用`routePath`        |

### 默认路由推导规则

如果页面没有显式声明`routePath`，宿主会根据文件路径推导。规则是：

- 取插件 ID。
- 取`frontend/pages/`后的相对路径。
- 将路径中的`/`替换为`-`。
- 将路径中的`_`替换为`-`。
- 最终拼成`<plugin-id>-<page-path>`。

例如：

| 文件路径                           | 推导结果                           |
| ---------------------------------- | ---------------------------------- |
| `frontend/pages/sidebar-entry.vue` | `plugin-demo-source-sidebar-entry` |
| `frontend/pages/user/profile.vue`  | `plugin-demo-source-user-profile`  |

### 页面开发约束

- 插件页面必须是**真实 Vue 源码文件**，而不是 JSON 描述。
- 页面内容应使用宿主已经公开的前端能力，不要直接依赖宿主未发布的内部实现。
- 如果页面需要请求插件自己的后端接口，建议接口路径保持清晰命名，例如`/plugins/<plugin-id>/summary`。

## 前端 `Slot` 接入

### 目录约定

插件`Slot`统一放在：

```text
frontend/slots/**/*.vue
```

推荐目录结构是“目录名即`slotKey`”，例如：

```text
frontend/slots/
  dashboard.workspace.after/
    workspace-card.vue
```

### `pluginSlotMeta`

`Slot`文件可以导出`pluginSlotMeta`：

```vue
<script lang="ts">
import { pluginSlotKeys } from "#/plugins/plugin-slots";

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceAfter,
};
</script>
```

当前支持字段如下：

| 字段       | 是否必填 | 说明                                      |
| ---------- | -------- | ----------------------------------------- |
| `pluginId` | 否       | 不传时默认从文件路径推导                  |
| `slotKey`  | 否       | 不传时默认从文件所在目录推导              |
| `order`    | 否       | 同一`Slot`下的排序值，越小越靠前，默认`0` |

### 默认 `slotKey` 推导规则

如果文件没有显式声明`slotKey`，宿主会读取其相对路径：

- 先去掉`frontend/slots/`前缀。
- 再去掉文件名。
- 剩余目录路径作为`slotKey`。

例如：

| 文件路径                                                      | 推导出的`slotKey`           |
| ------------------------------------------------------------- | --------------------------- |
| `frontend/slots/dashboard.workspace.after/workspace-card.vue` | `dashboard.workspace.after` |
| `frontend/slots/auth.login.after/login-tip.vue`               | `auth.login.after`          |

### 未发布插槽的处理方式

宿主只允许挂载已发布的`slotKey`。如果插件声明了未发布的插槽：

- 宿主会跳过该文件的挂载。
- 控制台会打印告警信息。
- 不会因为单个错误`Slot`影响其他页面或其他`Slot`。

### 当前已发布的前端插槽

| `slotKey`                      | 宿主位置         | 推荐用途               |
| ------------------------------ | ---------------- | ---------------------- |
| `auth.login.after`             | 登录页表单下方   | 提示信息、轻量入口     |
| `crud.table.after`             | 通用表格区域下方 | 说明卡片、辅助面板     |
| `crud.toolbar.after`           | 通用工具栏右侧   | 状态标签、快捷操作     |
| `dashboard.workspace.before`   | 工作台顶部       | 横幅、提醒、概览块     |
| `dashboard.workspace.after`    | 工作台底部       | 卡片、统计块、快捷入口 |
| `layout.header.actions.before` | 头部动作区前置   | 全局状态、入口         |
| `layout.header.actions.after`  | 头部动作区后置   | 快捷入口、轻量提示     |
| `layout.user-dropdown.after`   | 用户菜单左侧     | 轻量入口、状态提示     |

## SQL 约定

### 安装 SQL

插件安装 SQL 放在：

```text
manifest/sql/*.sql
```

规则如下：

| 规则       | 说明                                                |
| ---------- | --------------------------------------------------- |
| 文件名格式 | 必须是`{序号}-{当前迭代名称}.sql`                   |
| 序号格式   | 三位数字，例如`001`、`002`                          |
| 目录层级   | 必须直接位于`manifest/sql/`根目录，不能再嵌套子目录 |
| 扫描顺序   | 宿主按文件名排序后顺序执行                          |

### 卸载 SQL

插件卸载 SQL 放在：

```text
manifest/sql/uninstall/*.sql
```

规则如下：

| 规则       | 说明                                             |
| ---------- | ------------------------------------------------ |
| 文件名格式 | 与安装 SQL 相同                                  |
| 目录层级   | 必须直接位于`manifest/sql/uninstall/`根目录      |
| 发现方式   | 宿主在卸载流程中按目录约定单独发现               |
| 初始化隔离 | 宿主初始化流程不会扫描该目录，避免误执行卸载 SQL |

### 菜单与权限治理

菜单和权限相关信息必须遵循以下规则：

- 菜单、按钮权限和父子关系统一写在`plugin.yaml`的`menus`元数据中。
- 菜单稳定标识统一使用`sys_menu.menu_key`。
- 菜单父子关系通过`parent_key`解析真实`parent_id`。
- 不要在元数据或 SQL 中写死整型`id`或`parent_id`。
- 动态插件安装后由宿主按元数据幂等写入`sys_menu`，卸载时也由宿主按同一组菜单键删除菜单与角色关联。
- 插件 SQL 只保留业务表、业务种子或其他非菜单迁移，不再直接操作`sys_menu`和`sys_role_menu`。

换句话说：

- 菜单是否存在，以 manifest 菜单元数据为准。
- 插件是否启用，以插件治理状态为准。
- 页面文件是否可挂载，以前端源码文件和宿主运行时为准。

三者各自负责自己的真相源，不互相重复描述。

## 开发步骤

新增一个源码插件时，建议按以下顺序进行：

### 创建插件目录和模块

1. 在`apps/lina-plugins/`下创建`<plugin-id>/`目录。
2. 新建插件自己的`go.mod`。
3. 在根目录`go.work`中加入该插件模块路径。

### 编写最小清单

1. 新建`plugin.yaml`。
2. 只填写最小元数据。
3. 确认`id`使用`kebab-case`，并且与目录语义一致。

### 编写后端入口

1. 新建`backend/plugin.go`。
2. 调用`pluginhost.NewSourcePlugin("<plugin-id>")`。
3. 注册所需的路由、Hook 或其他扩展点。

### 更新宿主显式注册表

1. 修改`apps/lina-plugins/lina-plugins.go`。
2. 新增插件后端包的匿名导入。

### 编写前端页面和`Slot`

1. 页面放到`frontend/pages/`。
2. `Slot`放到`frontend/slots/`。
3. 需要显式元数据时分别导出`pluginPageMeta`和`pluginSlotMeta`。

### 编写 SQL

1. 安装 SQL 放到`manifest/sql/`。
2. 卸载 SQL 放到`manifest/sql/uninstall/`。
3. 只把业务表和业务数据迁移写进 SQL；插件菜单改在`plugin.yaml`的`menus`里声明。

### 验证

建议至少执行以下验证：

- `go test ./internal/service/plugin ./pkg/pluginhost`
- 插件相关的 E2E 用例
- 手工检查插件管理页、菜单显示、路由访问和禁用后的隐藏行为

## 开发约束

### 后端约束

- 插件后端代码应遵循宿主当前的`GoFrame`目录风格。
- `api/`和`internal/controller/`建议保持与宿主`gf gen ctrl`生成风格一致。
- 公开类型、结构体字段和方法应有足够英文注释，便于人工 review。
- 不要在插件里直接硬编码宿主未公开的内部包路径。

### 前端约束

- 页面和`Slot`必须是可直接参与宿主构建的真实 Vue 文件。
- 优先复用宿主已公开的组件和运行时能力。
- 不要依赖已经被删除的`pages.json`、`slots.json`或类似声明式文件。

### 元数据约束

- 只保留基础元数据。
- 不要把“约定可推导”的信息塞回`plugin.yaml`。
- 不要为了“配置更全”而重建已经被设计移除的模型。

## Review 清单

人工 review 一个源码插件时，建议按下面清单逐项确认：

| 检查项                                     | 结论标准                                                                           |
| ------------------------------------------ | ---------------------------------------------------------------------------------- |
| 插件目录位置是否正确                       | 位于`apps/lina-plugins/<plugin-id>/`                                               |
| 是否存在`go.mod`和`backend/plugin.go`      | `source`插件必须具备                                                               |
| `plugin.yaml`是否最小化                    | 不应再出现`schemaVersion`、`compatibility`、`entry`、`resources`、`metadata`等字段 |
| `id`是否唯一且符合`kebab-case`             | 宿主范围内唯一                                                                     |
| `lina-plugins.go`是否补了匿名导入          | 新插件必须显式接线                                                                 |
| 页面和`Slot`是否位于约定目录               | 页面在`frontend/pages/`，`Slot`在`frontend/slots/`                                 |
| 菜单和权限是否只在 manifest `menus` 中维护 | 不再通过插件 SQL 直接维护 `sys_menu/sys_role_menu`                                 |
| SQL 文件名和目录是否正确                   | 安装和卸载 SQL 分别放在正确目录，且文件名合规                                      |
| 禁用后是否能正确隐藏                       | 菜单、页面、`Slot`和路由都应受启停状态保护                                         |
| 文档是否足够清晰                           | 插件自身`README.md`应说明功能范围、路由、SQL 和验证方式                            |

## 常见错误

### 插件已写好，但插件管理页看不到

优先检查：

- `plugin.yaml`是否存在。
- `plugin.yaml`字段是否缺失。
- `id`是否与其他插件重复。

### 后端代码编译不过

优先检查：

- 是否创建了插件自己的`go.mod`。
- 根目录`go.work`是否已经包含该插件模块。
- `apps/lina-plugins/lina-plugins.go`是否已经追加匿名导入。

### 页面文件存在，但页面没有挂载

优先检查：

- 文件是否在`frontend/pages/`下。
- 组件是否存在默认导出。
- `pluginPageMeta.routePath`是否与菜单配置对应。

### `Slot`文件存在，但没有渲染

优先检查：

- 文件是否在`frontend/slots/`下。
- `slotKey`是否为宿主已发布的插槽。
- 插件当前是否已启用。

### 菜单存在，但访问返回 404

优先检查：

- 插件是否已启用。
- 后端路由是否通过`RegisterRoutes`正确注册。
- 路由是否挂到了期望的前缀下。

## 参考实现

当前仓库中最小可运行样例是`plugin-demo-source`：

| 文件                                                                    | 作用                     |
| ----------------------------------------------------------------------- | ------------------------ |
| `apps/lina-plugins/plugin-demo-source/plugin.yaml`                      | 源码插件最小清单示例     |
| `apps/lina-plugins/plugin-demo-source/backend/plugin.go`                | 源码插件后端注册入口示例 |
| `apps/lina-plugins/plugin-demo-source/frontend/pages/sidebar-entry.vue` | 源码插件页面示例         |
| `apps/lina-plugins/plugin-demo-dynamic/plugin.yaml`                     | 动态插件最小清单示例     |
| `apps/lina-plugins/plugin-demo-dynamic/plugin_embed.go`                 | 动态插件作者侧资源声明示例 |
| `apps/lina-plugins/plugin-demo-dynamic/backend/api/`                    | 动态插件路由合同定义示例 |
| `apps/lina-plugins/plugin-demo-dynamic/main.go`                         | 动态插件`Wasm bridge`入口示例   |
| `apps/lina-plugins/plugin-demo-dynamic/frontend/pages/mount.js`         | 动态内嵌挂载入口示例     |
| `apps/lina-plugins/plugin-demo-dynamic/frontend/pages/standalone.html`  | 动态独立静态页示例       |

如果要新增新插件，建议先复制`plugin-demo-source`的整体结构，再按本文档约束删减或扩展，而不是从零随意拼目录。
