## 0. 当前实现快照（2026-04-06）

> 当前仓库已完成**第一期：源码插件底座**。本节用于说明已交付范围；后续 checklist 继续作为 `package/wasm`、多节点热更新与开发者工具的路线图。

- [x] 0.1 新增 `apps/lina-plugins/plugin-demo/` 示例目录、`plugin.yaml` 与插件资源约定
- [x] 0.2 新增 `sys_plugin` 并仅为宿主 `sys_*` 表生成 DAO/DO/Entity，插件私有表不进入宿主生成产物
- [x] 0.3 实现源码插件扫描、注册表同步、启用/禁用 API 与插件管理页，源码插件默认作为随宿主编译的已集成插件管理
- [x] 0.4 实现 `plugin-demo` 最小回调注册式扩展示例、插件页面/Slot 挂载与菜单联动隐藏，且插件特定前后端实现收敛在 `apps/lina-plugins/plugin-demo/`
- [x] 0.5 新增并补齐 `TC0066-source-plugin-lifecycle` 与插件管理 POM，覆盖 source plugin 的 sync/enable/disable、编译整合与 slot 渲染/隐藏
- [x] 0.6 完成 source plugin 的免安装闭环、首批通用 Hook/Slot 与一期验收收口；后续进入 `package/wasm` 与多节点热更新阶段

## 第一期当前落地快照（2026-04-06）

- [x] 建立 `apps/lina-plugins/<plugin-id>/` 目录规范，并要求 `plugin-demo` 的插件特定前后端实现收敛在插件目录维护
- [x] 落地源码插件发现、同步、启用/禁用、菜单隐藏与后端扩展点示例，源码插件默认不走安装/卸载流程
- [x] 将 `plugin-demo` 后端示例收敛到插件目录内的 Go 源码实现，并通过构建期静态注册表接入宿主
- [x] 提供 `plugin-demo` 前端页面与 Slot 源码，并通过宿主通用运行时页/Slot 装载器挂载
- [x] 补齐源码插件免安装管理闭环
- [x] 基于插件目录后端注册与 `frontend/slots/**/*.vue` 抽象首批通用 Hook/Slot 总线

## 1. 契约与元数据底座

- [ ] 1.1 定义 `plugin.yaml` 清单 schema、版本策略、兼容矩阵与宿主校验流程
- [ ] 1.2 规划 `apps/lina-plugins/<plugin-id>/` 标准目录结构，并补齐源码插件脚手架模板
- [ ] 1.3 新增插件元数据 SQL 方案，落地 `sys_plugin`、`sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state` 等基础表
- [ ] 1.4 基于新增表生成 DAO/DO/Entity，并建立插件注册、生命周期、资源引用、迁移记录的后端服务骨架
- [ ] 1.5 定义插件管理后台 API、DTO、管理页面信息结构以及状态机枚举

## 2. 第一期：源码插件接入

- [x] 2.1 实现源码插件扫描与后端注册表同步，确保新增插件目录后无需手工修改核心装配代码
- [x] 2.2 实现前端源码插件清单生成、页面入口发现、Slot 注册与宿主构建集成
- [x] 2.3 打通源码插件的同步发现、启用、禁用管理流程和后台管理界面；运行时安装/卸载留给 `package/wasm`
- [x] 2.4 实现 `plugin-demo` 源码插件后端能力，覆盖插件目录 Go 源码接入、公开/受保护路由与治理接入
- [x] 2.5 实现 `plugin-demo` 源码插件前端能力，覆盖菜单页展示、宿主页面接入与基本管理交互

## 3. 第一期：治理接入与扩展点发布

- [x] 3.1 扩展菜单、角色与权限链路，使插件菜单和插件权限复用 Lina 通用治理模块
- [x] 3.2 建立宿主后端 Hook 总线，发布首批认证与插件生命周期 Hook，并实现失败隔离与执行观测
- [x] 3.3 建立宿主前端 Slot 注册表，发布首批布局与工作台 Slot，并实现加载失败降级机制
- [x] 3.4 完成插件禁用、重启用及运行时插件卸载时的菜单隐藏、权限失效、角色关系保留与资源清理联动

## 4. 第二期：运行时 package 与 wasm 插件

- [ ] 4.1 定义运行时 `package` 格式、`wasm-only` 兼容规则、资源嵌入约定与 ABI 版本策略
- [ ] 4.2 实现运行时插件安装器、校验器、资源提取器与迁移执行器
- [ ] 4.3 基于 WASM Runtime 实现插件加载、Hook 调用、超时控制、错误隔离与卸载回收
- [ ] 4.4 实现运行时插件静态资源托管与三种前端接入模式：`iframe`、新标签页、宿主内嵌挂载
- [ ] 4.5 让 `plugin-demo` 同时产出 `package` 与 `wasm` 运行时版本，并验证与源码版本契约一致

## 5. 第三期：多节点热更新与回滚

- [ ] 5.1 建立插件 `desired_state/current_state/generation/release_id` 代际模型与主节点切换流程
- [ ] 5.2 将主节点选举与节点 Reconciler 接入插件安装、启停、升级与状态收敛链路
- [ ] 5.3 实现插件热升级时的新旧代际切换、旧请求自然结束与节点状态上报
- [ ] 5.4 实现插件升级失败回滚、迁移异常恢复与前端资源切换失败保护机制
- [ ] 5.5 实现前端插件代际感知与“当前插件页面刷新提示”，保证非插件页面用户无感

## 6. 文档、模板与开发者工具

- [ ] 6.1 编写插件开发指南，覆盖 `source`、`package`、`wasm` 三种模式的目录、清单、权限、菜单和扩展点约定
- [ ] 6.2 编写插件运维指南，覆盖安装、启停、卸载、升级、回滚、多节点注意事项与故障排查
- [ ] 6.3 提供插件模板与打包脚本，帮助开发者快速创建源码插件和运行时产物
- [ ] 6.4 补充 `plugin-demo` 的设计说明、发布说明与宿主接入说明，作为后续插件开发参考样板

## 7. E2E 与验收验证

- [x] 7.1 完成 `hack/tests/e2e/plugin/TC0066-source-plugin-lifecycle.ts`，覆盖源码插件 `sync/enable/disable`、编译整合与工作台 slot 渲染切换
  - [x] TC-66a：同步 source 插件后自动处于已集成态，插件管理页无安装按钮
  - [x] TC-66b：启用插件后渲染工作台 slot，并展示左侧插件菜单页
  - [x] TC-66c：启用后可验证插件路由与鉴权访问控制
  - [x] TC-66d：禁用后隐藏工作台 slot，并隐藏插件菜单
  - [x] TC-66e：禁用后源码插件仍保留已集成态，且无需重新安装
- [ ] 7.2 创建 `hack/tests/e2e/plugin/TC0067-runtime-package-lifecycle.ts`，覆盖 `package` 插件安装、启停、卸载与资源托管
- [ ] 7.3 创建 `hack/tests/e2e/plugin/TC0068-runtime-wasm-lifecycle.ts`，覆盖 `wasm` 插件安装、启停、失败隔离与回滚
- [ ] 7.4 创建 `hack/tests/e2e/plugin/TC0069-plugin-permission-governance.ts`，覆盖角色授权、菜单可见性、权限恢复与数据权限上下文
- [ ] 7.5 创建 `hack/tests/e2e/plugin/TC0070-plugin-hot-upgrade.ts`，覆盖热升级、当前页面刷新提示、多节点代际切换与回退
- [x] 7.6 为插件管理与插件页面补充所需的 POM（安装/卸载、slot 可见性断言），保证 `TC0066` 可独立运行

## Feedback

- [x] **FB-1**: `gf gen dao` 只处理宿主 `sys_*` 数据表，插件私有 `plugin_*` 表不再生成到 `lina-core` 的 DAO/DO/Entity
- [x] **FB-2**: 合并 `011-plugin-framework.sql` 与 `012-plugin-lifecycle-state.sql`，同一迭代只保留 1 个 SQL 文件
- [x] **FB-3**: 在项目开发规范文档中明确“宿主 `manifest/sql/` 目录下同一迭代只保留 1 个版本 SQL 文件”
- [x] **FB-4**: 精简 `011-plugin-framework.sql` 的表结构变更逻辑，插件一期按新功能处理，仅保留 `CREATE TABLE`，去掉冗余结构 SQL
- [x] **FB-5**: 插件 SQL 采用与宿主一致的版本命名；卸载 SQL 独立放到 `manifest/sql/uninstall/`，避免被初始化顺序执行误扫
- [x] **FB-6**: `plugin.yaml` 作为统一入口索引菜单声明；插件菜单改用 `sys_menu.menu_key` 与 `parent_key` 维护，去掉对 `remark` 和固定整型 `id/parent_id` 的依赖
- [x] **FB-7**: 未交付阶段将 `sys_menu` 的 `menu_key` 结构与宿主插件菜单种子回收到 `008-menu-role-management.sql`，移除 `011-plugin-framework.sql` 中对应冗余 SQL
- [x] **FB-8**: `plugin-demo` 安装 SQL 去掉 `UPDATE/ON DUPLICATE KEY UPDATE`，插件菜单与授权种子统一使用 `INSERT IGNORE INTO` 幂等写入
- [x] **FB-9**: 删除 `plugin-demo` 冗余的 `manifest/menus.json` 与 `resources.menus` 索引，插件一期菜单以 SQL 为单一真相源
- [x] **FB-10**: 源码插件改为随宿主编译即集成，插件管理页不再为 `source` 类型展示安装/卸载按钮，源码插件默认视为已集成
- [x] **FB-11**: 支持插件目录内后端 Go 源码随宿主一起编译接入，并用 `plugin-demo` 走通“开发-编译-展示”完整链路
- [x] **FB-12**: 调整 `TC0066-source-plugin-lifecycle`，改为验证源码插件“同步发现 + 启用/禁用 + 编译接入后的后端扩展点行为”闭环
- [x] **FB-13**: 修复 `make dev` 后端进程后台保活问题，保证源码插件一期“开发-编译-展示”链路可稳定验证
- [x] **FB-14**: 调整 `plugin-demo` 插件首页体验，菜单打开后台 Tab 页后展示更直观的示例信息，明确告知插件已生效
- [x] **FB-15**: 源码插件首次同步后默认启用，且后续同步不覆盖管理员显式禁用状态
- [x] **FB-16**: `plugin-demo` 需提供“左侧主菜单顶部入口 + 右上角菜单栏入口”两个插件示例页面，并均以内页 Tab 方式打开
- [x] **FB-17**: 插件管理页类型展示调整为“源码插件 / 运行时插件”，并将 `package`、`wasm` 作为运行时交付方式展示
- [x] **FB-18**: 清理 `plugin-demo` 前端重复实现，仅保留当前真实生效的页面/Slot 源码资源
- [x] **FB-19**: 修复已启用 `plugin-demo` 后左侧插件菜单未展示的问题，并验证菜单可见性与排序
- [x] **FB-20**: 修复右上角“插件示例”入口点击后 404 的问题，并验证入口以内页 Tab 方式正确打开
- [x] **FB-21**: 修复按钮类型菜单被错误返回到左侧导航/动态路由中的问题，并验证按钮权限不再显示为可导航菜单
- [x] **FB-22**: 修复左侧菜单未按菜单管理排序规则展示的问题，并验证同级菜单按排序号稳定输出
- [x] **FB-23**: 修复“插件管理”被展示为独立顶级目录的问题，将其调整为“系统管理”下的直属菜单
- [x] **FB-24**: 修复页面刷新时重复出现两个“加载菜单中”提示的问题，并验证首次加载只触发一次菜单装载提示
- [x] **FB-25**: 修复排序号为 `0` 的顶级菜单在动态路由响应中丢失 `order` 字段，导致“仪表盘”被前端排到菜单底部
- [x] **FB-26**: 将一期源码插件前端从 `pages/slots/*.json` 描述切换为真实前端源码文件实现，并验证 `plugin-demo` 页面与 Slot 仍可正常挂载
- [x] **FB-27**: 简化 `plugin-demo` 示例插件，移除右上角菜单/页面与登录审计页面入口，仅保留左侧菜单页并收敛其展示内容
- [x] **FB-28**: 补齐宿主系统菜单初始化种子数据的 `menu_key` 字段值，避免 `008-menu-role-management.sql` 初始化后出现空业务标识
- [x] **FB-29**: 在无历史数据债务前提下，直接修改 `008-menu-role-management.sql` 原始菜单种子 `INSERT`，为每条宿主菜单显式写入 `menu_key`，并移除初始化后的回填 `UPDATE`
- [x] **FB-30**: 调整宿主系统菜单的 `menu_key` 命名，移除 `host:` 前缀，仅保留插件菜单使用 `plugin:` 命名空间，并避免宿主插件管理菜单与插件命名空间冲突
- [x] **FB-31**: 清理 `plugin-demo` 清单中的冗余 `backend.apis` 声明，并同步更新 README，明确源码插件后端能力由 Go 编译期注册而非 `plugin.yaml` 路由声明驱动
- [x] **FB-32**: 补齐插件前后端插槽目录、类型化安装位置定义与开发者技术文档，禁止在宿主与插件示例中硬编码 Hook/Slot 位置字符串
- [x] **FB-33**: 将源码插件后端扩展模型升级为回调注册式宿主扩展点，补齐 `auth.login.failed`、`http.route.register`、`http.request.after-auth`、`cron.register`、`menu.filter`、`permission.filter`
- [x] **FB-34**: 在布局、登录页、工作台与 CRUD 通用壳层补齐更多前端 Slot，避免扩展点过度集中在少数页面
- [x] **FB-35**: 更新 `plugin-demo` 示例插件，覆盖新的回调注册后端能力与新增前端 Slot 示例
- [x] **FB-36**: 更新 `apps/lina-plugins/README.md` 与相关插件示例文档，明确新的扩展点目录、注册方式与推荐用法
- [x] **FB-37**: 扩展 `TC0066-source-plugin-lifecycle`，验证新增通用扩展点的可见性、路由装配与鉴权后回调行为
- [x] **FB-38**: 统一后端事件 Hook 与注册式回调扩展点的 Go 类型常量目录，禁止宿主与插件代码继续硬编码后端扩展点字符串
- [x] **FB-39**: 为后端回调注册接口补齐执行模式参数，区分 `blocking` 与 `async`，并由宿主校验每个扩展点允许的执行模式
- [x] **FB-40**: 删除 `internal/service/plugin/plugin.go` 内对 Hook 常量的重复别名导出，并同步更新 `plugin-demo` 与 `apps/lina-plugins` 文档示例
- [x] **FB-41**: 调整 `pkg/pluginhost` 包内源码文件命名，统一使用 `pluginhost_*.go` 风格并同步更新文档引用
- [x] **FB-42**: 去掉后端扩展点公开类型与常量中的 `Backend` 前缀，统一收敛为 `ExtensionPoint*` 风格命名
- [x] **FB-43**: 为插件定时任务注册输入对象补齐“当前是否主节点”的识别方法，供插件自行决定主节点专属执行逻辑
- [x] **FB-44**: 将插件回调注册接口中的对象型输入参数抽象为公开接口，降低插件对宿主内部结构体的直接耦合
- [x] **FB-45**: 精简 `plugin-demo` 后端示例，移除登录审计数据库演示代码，为关键逻辑补充注释，并将示例定时任务调整为每分钟执行
- [x] **FB-46**: 同步更新 `apps/lina-plugins` 与 `plugin-demo` 文档、清单及 E2E 用例，匹配新的命名、接口契约与示例行为
- [x] **FB-47**: 将 `plugin-demo` 子目录说明文档收敛到根目录 `README.md`，移除 `backend`、`frontend`、`manifest` 下的介绍性 `README.md`
- [x] **FB-48**: 将源码插件前端目录约定从 `frontend/src/pages|slots` 统一收敛为 `frontend/pages|slots`，同步宿主扫描逻辑、示例文件与文档规格
- [x] **FB-49**: 移除 `plugin-demo` 的 `crud.table.after` 前端示例，避免在所有基于 `useVbenVxeGrid` 的页面下方默认展示说明内容
- [x] **FB-50**: 修复菜单管理页在树表自动高度场景下页面高度持续增长的问题，并补齐页面高度稳定性的回归验证
- [x] **FB-51**: 将列表页高度持续增长问题收敛到 `useVbenVxeGrid` 共享层修复，并补齐菜单管理、角色管理、操作日志页面的高度稳定回归验证
- [x] **FB-52**: 修复页面重新获得焦点时插件注册表无条件触发菜单与路由重算的问题，避免左侧菜单在插件状态未变化时闪烁重渲染
- [x] **FB-53**: 保持 `plugin-demo` 左侧菜单页原有正文内容不变，修复打开页面时错误弹出的“最小源码插件接入”提示框
- [x] **FB-54**: 简化 `plugin-demo` 左侧菜单页内容，仅保留标题和一句简要介绍
- [x] **FB-55**: 让 `plugin-demo` 页面展示的一小部分内容来自插件后端接口，并将插件后端路由注册升级为与宿主一致的对象式 `Bind` 管理方式
- [x] **FB-56**: 收敛 `plugin-demo` 摘要接口，只保留页面实际使用的 `message` 字段，并同步清理冗余 DTO、service 输出与 E2E 断言
- [x] **FB-57**: 将 `plugin-demo` 后端新增对象式路由模块命名从 `example` 统一收敛为 `demo`，保持插件名与源码模块命名一致
- [x] **FB-58**: 插件 HTTP 路由注册改为宿主独立无前缀分组，并向插件公开宿主可选中间件目录，由插件自行决定路由前缀与中间件组合
- [x] **FB-59**: 明确并示范插件可拆分多个治理策略不同的路由分组进行注册，支持同一插件同时暴露免鉴权与需鉴权接口
- [x] **FB-60**: 规范源码插件后端 `api/controller` 目录命名，要求与宿主 GoFrame `gf gen ctrl` 生成风格保持一致
- [x] **FB-61**: 按 `gf gen ctrl` 实际生成结果重整 `plugin-demo` 控制器目录，拆分 public/protected API 模块并删除旧的手写控制器目录
- [x] **FB-62**: 精简插件路由注册契约，移除 `Public/Protected` 封装分组，仅保留根分组与宿主已发布中间件目录供插件自行组合
- [x] **FB-63**: 插件路由注册支持宿主同款 `group.Group(..., func(*ghttp.RouterGroup){...})` 风格，并将 `plugin-demo` 收敛回单一 `demo` API 模块，通过方法级绑定演示匿名与鉴权路由
- [x] **FB-64**: 将插件路由注册输入对象 `RouteRegistrars` 收敛为单数命名 `RouteRegistrar`，并同步更新构造函数、调用点与开发文档
- [x] **FB-65**: 精简插件管理页列表字段，移除“交付方式/接入态/入口”列，新增“安装时间”并将“类型”列标题调整为“插件类型”
- [x] **FB-66**: 调整插件管理页列表列顺序，将“描述”列移动到“版本”和“状态”列之间
- [x] **FB-67**: 修复插件管理页描述列悬停时出现重复提示的问题，仅保留单一提示层
- [x] **FB-68**: 将插件管理页描述提示改为页面内单例自定义浮层，彻底关闭组件库与表格自身的额外提示
- [x] **FB-69**: 修复插件管理页描述列 hover 数秒后仍出现浏览器原生提示的问题，仅保留页面自定义 tooltip
- [x] **FB-70**: 移除插件管理页描述列自定义 tooltip，改为仅保留单一系统默认提示
- [x] **FB-71**: 精简 `plugin-demo` 后端示例，移除 `RegisterAfterAuthHandler` 与 `RegisterCron` 注册，并同步更新文档与回归用例
