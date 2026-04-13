## 1. 构建约定

- [x] 1.1 为动态插件定义统一的 `go:embed` 资源声明约定，并在样例插件中提供根级资源声明文件
- [x] 1.2 调整 `hack/build-wasm` 资源收集流程，优先消费动态插件声明的嵌入文件系统
- [x] 1.3 保留目录扫描回退逻辑，确保未迁移的动态插件仍可继续构建

## 2. 运行时快照

- [x] 2.1 保持动态插件 manifest、前端资源、安装 SQL、卸载 SQL 等自定义节快照格式兼容
- [x] 2.2 校验宿主上传、装载、启用与前端托管链路在新构建产物下无需引入 guest 资源读取逻辑
- [x] 2.3 更新 `plugin-demo-dynamic` 样例和相关说明，明确作者侧统一声明、宿主侧快照治理的边界

## 3. 验证与文档

- [x] 3.1 为构建器补充或更新测试，覆盖嵌入资源声明和目录扫描回退两条路径
- [x] 3.2 运行动态插件相关构建与测试，确认新旧构建路径产物均可被宿主正确解析
- [x] 3.3 更新插件开发文档，说明动态插件推荐使用 `go:embed` 声明资源以及当前兼容策略

## Feedback

- [x] **FB-1**: 动态插件构建产物和 guest runtime 中间 `wasm` 不应再写回 `apps/lina-plugins/<plugin-id>/temp/`，而应统一收敛到仓库根 `temp/output/`
- [x] **FB-2**: 将 `apps/lina-core/internal/service/plugin` 中与业务无关的插件资源文件系统能力抽离到 `apps/lina-core/internal/pkg/`，降低 `plugin` 组件复杂度
- [x] **FB-3**: 按职责拆分 `hack/build-wasm/builder/builder.go`，避免单文件维护成本继续上升
- [x] **FB-4**: 对齐 `apps/lina-core/pkg/pluginbridge` 与 `pluginhost` 的文件命名规范，统一使用组件前缀命名
- [x] **FB-5**: 将已抽离的 `pluginfs` 通用能力从 `apps/lina-core/internal/pkg/` 提升到 `apps/lina-core/pkg/`，确保公共能力可被组件外复用
- [x] **FB-6**: 修正 `hack/build-wasm/builder` 的目录嵌入扫描逻辑，确保与 `go:embed` 对隐藏文件和下划线文件的默认过滤语义一致
- [x] **FB-7**: 修正 `hack/build-wasm/builder` 在无 `go.mod` 插件目录构建 guest runtime 时遗留临时模块文件的问题
- [x] **FB-8**: 为 `hack/build-wasm/builder` 拆分后的实现文件补齐符合规范的文件顶部用途注释
- [x] **FB-9**: 为 `plugin_source_embedded` 与 `pluginbridge` 相关文件补齐或修正顶部用途注释格式，满足文件注释规范
- [x] **FB-10**: 修正 `pkg/logger` 与 `pkg/pluginbridge` 组件文件顶部注释，统一满足主文件与实现文件的注释格式规范
