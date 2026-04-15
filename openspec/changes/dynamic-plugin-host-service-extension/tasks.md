## 1. 协议与运行时骨架

- [x] 1.1 在`pluginbridge`中定义结构化宿主服务调用 envelope、协议版本和统一错误模型。
- [x] 1.2 将`lina_env.host_call`重构为统一宿主服务调用通道，移除按能力堆叠专用 opcode 的公开协议设计。
- [x] 1.3 在动态插件运行时中实现宿主服务注册表、统一分发器和执行上下文装配。
- [x] 1.4 为宿主服务调用协议、运行时服务分发和错误模型补充单元测试。

## 2. 清单、产物与治理集成

- [x] 2.1 扩展动态插件`plugin.yaml`与构建器，支持`hostServices`声明和静态校验。
- [x] 2.2 将宿主服务治理快照写入`wasm`自定义节，并在运行时产物解析链路中恢复。
- [x] 2.3 扩展`sys_plugin_resource_ref`同步逻辑，统一纳入存储、上游、数据表授权投影、缓存、锁、密钥、事件主题、队列和通知通道等治理目标。
- [x] 2.4 为 manifest／artifact 校验和资源归属同步补充治理测试。

## 3. 首批宿主服务实现

- [x] 3.1 实现`runtime` service，统一承载日志、状态和宿主基础信息能力。
- [x] 3.2 实现`storage` service，支持逻辑存储空间的`put`、`get`、`delete`、`list`和`stat`能力。
- [x] 3.3 为`storage` service 接入大小限制、类型限制、公开性治理和隔离校验。
- [x] 3.4 实现`network` service，支持基于授权 URL 模式的同步 HTTP 请求。
- [x] 3.5 为`network` service 接入平台级头部保护、默认 timeout 和最大响应体限制。
- [x] 3.6 实现`data` service，支持基于`table`和宿主 DAO / `gdb` ORM 封装的查询、详情、新增、更新、删除和事务边界。
- [x] 3.7 为`data` service 接入当前用户权限、数据范围注入、`DoCommit` 宿主拦截、审计能力，并禁止 raw SQL 公开能力。
- [x] 3.8 抽离`pkg/plugindb/host`，统一承载`data service`自定义`gdb` Driver / DB wrapper、审计上下文和`DoCommit`治理骨架。

## 4. Guest SDK 与样例

- [x] 4.1 为 guest 侧提供`runtime`、`storage`、`network`和`data`宿主服务 SDK wrappers。
- [x] 4.2 更新动态插件样例，覆盖文件、网络和数据访问场景。
- [x] 4.3 更新插件开发文档，明确宿主能力统一通过结构化宿主服务获取。
- [x] 4.4 为统一宿主服务模型补充运行时集成测试。
- [x] 4.5 新增`pkg/plugindb/shared`强类型 query plan、事务操作和排序/过滤枚举模型，禁止在实现中直接使用枚举语义字符串字面量。
- [x] 4.6 新增`pkg/plugindb`受限 ORM 风格 guest SDK，对插件作者暴露`Open().Table(...).WhereEq(...).All()`等链式数据访问体验，同时底层继续走受治理的 hostService。
- [x] 4.7 将动态插件数据访问推荐路径迁移到`pkg/plugindb`，同步更新 demo、开发文档和样例代码；`pluginbridge.Data()`降级为兼容层。
- [x] 4.8 为`pkg/plugindb`的 shared / guest / host 初始实现补充单元测试。

## 5. 低优先级宿主服务实现

- [ ] 5.1 实现`cache` service，支持命名缓存空间的`get`、`set`、`delete`、`incr`和`expire`能力。
- [ ] 5.2 实现`lock` service，支持命名锁资源的`acquire`、`renew`和`release`能力。
- [ ] 5.3 实现`secret` service，支持密钥引用解析和最小暴露控制。
- [ ] 5.4 实现`event` service，支持命名事件主题的`publish`能力。
- [ ] 5.5 实现`queue` service，支持命名队列的`enqueue`能力。
- [ ] 5.6 实现`notify` service，支持命名通知通道的`send`能力。
- [ ] 5.7 为`cache`、`lock`、`secret`、`event`、`queue`、`notify`补充宿主授权、限额和审计测试。

## 6. E2E 验证

- [x] 6.1 创建`hack/tests/e2e/plugin/TC0071-runtime-wasm-host-services.ts`。
- [x] 6.2 实现`TC-71a`：已授权的`runtime`、`storage`、`network`和`data`宿主服务调用成功。
- [x] 6.3 实现`TC-71b`：未声明 service、method 或未授权资源标识（`resourceRef` / `table`）的调用被宿主拒绝。
- [x] 6.4 实现`TC-71c`：插件尝试申请 raw SQL 或未授权宿主能力时被宿主拒绝。
- [x] 6.5 创建`hack/tests/e2e/plugin/TC0073-plugin-host-service-authorization-review.ts`。
- [x] 6.6 实现`TC-73a~c`：安装与启用弹窗展示申请权限，并持久化最终授权结果。
- [ ] 6.7 创建`hack/tests/e2e/plugin/TC0072-runtime-wasm-host-services-low-priority.ts`。
- [ ] 6.8 实现`TC-72a`：已授权的`cache`、`lock`、`secret`、`event`、`queue`、`notify`宿主服务调用成功。
- [ ] 6.9 实现`TC-72b`：低优先级宿主服务在未授权资源或超限场景下被宿主拒绝。

## Feedback

- [x] **FB-1**: `data service` 必须通过宿主 DAO / `gdb` ORM 契约执行，禁止 guest 侧 raw SQL / 通用 SQL 执行设计。
- [x] **FB-2**: `data service` 需要在 GoFrame `DoCommit` 提交链路建立宿主拦截点，用于权限控制、事务治理和审计。
- [x] **FB-3**: 补齐`runtime`、`storage`、`network`、`data`四类已实现 hostServices 的单元测试覆盖，补充授权拒绝、治理分支和运行时状态校验。
- [x] **FB-4**: 创建`hack/tests/e2e/plugin/TC0071-runtime-wasm-host-services.ts`，覆盖核心 hostServices 的成功调用与未授权拒绝场景。
- [x] **FB-5**: 补充 raw SQL 能力在构建链路与动态插件上传链路中的拒绝测试，确保旧能力模型不会重新暴露。
- [x] **FB-6**: 将所有带`resourceRef`的 hostServices 统一为“声明即申请，安装/启用时由宿主确认授权”的治理模型，并在安装/启用流程展示申请权限与最终授权结果。
- [x] **FB-7**: 将`data service`简化为表级授权模型；运行时按表名直接授权和访问，移除`resourceRef`／命名数据资源层。
- [x] **FB-8**: 将`data service`的表申请重新收敛到统一`resources`结构下，改为`resources.tables`，以保持资源声明外形一致并为后续扩展更多 data 资源类型预留空间。
- [x] **FB-9**: 简化`network service`的清单声明模型，仅声明 URL 模式；宿主授权后按 URL 模式匹配放行，不再要求插件声明方法白名单、头白名单和独立上游引用名。
- [x] **FB-10**: 明确`network` URL pattern 的匹配规则与边界，包括 scheme、host 通配、path 前缀、query/fragment 处理和默认拒绝策略，补充到 spec/design 方便最终审查。
- [x] **FB-11**: 将`storage service`从`resourceRef + attributes`收敛为`resources.paths`模型，仅声明授权逻辑路径或路径前缀，并在 spec/design 中明确路径归一化、前缀匹配与默认拒绝规则。
- [x] **FB-12**: 将`storage service`的 manifest、授权快照、运行时匹配、guest SDK、demo 与测试全面收敛为`resources.paths`模型，移除公开的`resourceRef + attributes`依赖。
- [x] **FB-13**: 将`data service`的读取型 method 从`query`重命名为`list`，与`get`形成“列表查询 / 单条获取”的清晰分工，并同步更新协议、SDK、demo、spec 与测试。
- [x] **FB-14**: 将`apps/lina-plugins/README.md`中动态插件宿主服务相关开发说明同步到最新 data hostService 命名，明确`host:data:read`对应`list/get`、`host:data:mutate`对应写操作，并强调不暴露 raw SQL。
- [x] **FB-15**: 将`apps/lina-plugins/plugin-demo-dynamic/README.md`中的 data hostService 示例与说明同步为最新命名，明确`host:data:read`与`list/get`分工，并强调只允许表级结构化访问。
- [x] **FB-16**: 修复`pnpm -C apps/lina-vben -F @lina/web-antd typecheck`当前已有的前端类型错误，补齐页面、表格、用户态与组件签名的类型约束，恢复主应用可通过类型检查。
- [x] **FB-17**: 将动态插件`data service`的 guest 侧 API 从简单表操作封装演进为`pkg/plugindb`受限 ORM 风格 facade，保持底层 ABI 仍为结构化 hostService，避免直接暴露完整`gdb`/DAO 能力。
- [x] **FB-18**: 为`plugindb`中的查询动作、过滤操作符、排序方向、事务 mutation 类型和访问模式等枚举语义值定义独立 Go 命名类型与常量，并禁止在 builder、query plan、执行器和审计逻辑中直接写字符串字面量。
- [x] **FB-19**: 将`data service`当前位于`internal/service/plugin/internal/datahost`中的自定义 Driver / DB wrapper / 审计上下文能力上提到`pkg/plugindb/host`，形成宿主可复用治理层。
- [x] **FB-20**: 将动态插件 demo 与数据访问开发文档迁移为`plugindb.Open()`主路径，并补充从兼容层`pluginbridge.Data()`向`plugindb`过渡的说明。
- [x] **FB-21**: 为动态插件`plugin.yaml`中本次迭代新增的`hostServices`及其资源声明字段补齐就地注释说明，确保样例清单可直接作为作者侧参考模板。
- [x] **FB-22**: 将动态插件 demo 控制器中的业务负载逻辑下沉到`backend/internal/service`组件，保持控制器仅负责桥接请求与响应装配。
- [x] **FB-23**: 更新动态插件生命周期 E2E 辅助逻辑，使启用带 hostServices 授权弹窗的动态插件时能够按默认授权流继续执行回归验证。
- [x] **FB-24**: 移除`pkg/pluginhost`中已无源码插件使用的`ResourceSpec`及其 source-plugin 适配链，避免继续保留与当前 data hostService / backend resource 模型不一致的冗余结构。
- [x] **FB-25**: 为宿主服务授权匹配、数据访问上下文校验等复杂治理逻辑补充解释性注释，说明实现思路与关键分支，降低后续维护理解成本。
- [x] **FB-26**: 将插件运行时中的授权状态、执行来源等枚举语义字符串收敛为独立命名类型与常量，避免继续在运行时、数据治理和审计链路中硬编码字符串字面量。
- [x] **FB-27**: 按`goframe-v2`代码风格要求收敛本次实现中的相关多变量连续定义，使用`var` block 提升可读性与维护一致性。
- [x] **FB-28**: 清理已不可达的动态插件 raw SQL 旧 hostcall 链路（`host:db:*` 编解码与宿主处理器），避免继续保留与当前结构化 hostService 模型不一致的冗余代码。
- [x] **FB-29**: 收敛插件生命周期 facade 中`Install`/`InstallWithAuthorization`及启停相关公开方法的重复包装关系，改为复用私有辅助流程，避免 exported 方法互相转调。
- [x] **FB-30**: 将动态插件作者侧 manifest 与 runtime artifact 收敛为仅声明`hostServices`，移除顶层`capabilities`作者输入与产物自定义节，宿主内部 capability 分类改为从`hostServices.methods`自动推导。
- [x] **FB-31**: 将动态插件样例、guest SDK 与相关测试 fixture 中新增或维护的错误创建统一改为`gerror`，并为关键失败分支补充上下文包装说明。
- [x] **FB-32**: 将插件生命周期 facade 的安装与状态切换公开 API 进一步收敛为单一入口（通过可空授权参数表达是否附带授权确认），并审查同类重复包装点，仅保留真正有语义价值的快捷方法。
- [x] **FB-33**: 按项目 Go 文件注释规范修正`apps/lina-core/pkg/plugindb`及其子包源码文件头，确保主文件与非主文件的注释职责、空行位置和文件用途说明一致。
- [x] **FB-34**: 将动态插件 guest 侧 `pluginbridge` 宿主服务 client 的公开返回值收敛为接口类型，并同步调整 demo 与相关调用方，避免继续向插件作者暴露具体实现结构体。
- [x] **FB-35**: 对`sys_plugin_resource_ref`相关实现与文档补充语义澄清，明确其是插件治理资源索引，而非仅服务于`resourceRef`字段的镜像表。
- [x] **FB-36**: 将动态插件 demo 中稳定 JSON 响应负载的`map[string]any`实现收敛为结构体模型，减少硬编码键名并提升可维护性。
- [x] **FB-37**: 为 data hostService 授权展示补充数据表的人类可读说明（优先展示宿主表注释），同步更新安装/启用授权弹窗与`TC0073`回归覆盖。
- [x] **FB-38**: 收敛`apps/lina-core/internal/service/plugin/internal/integration/resource_ref.go`中散落的稳定治理标识与重复文案硬编码，统一为常量和辅助构造函数，降低后续维护成本。
