## ADDED Requirements

### Requirement: 宿主发布稳定的后端 Hook 契约
系统 SHALL 发布一组具名、版本化、可审计的后端 Hook，供插件在宿主关键业务事件上扩展行为。

#### Scenario: 登录成功事件触发 Hook
- **WHEN** 用户登录成功且宿主发布 `auth.login.succeeded` Hook
- **THEN** 宿主按约定上下文向已启用插件分发该事件
- **AND** 上下文至少包含用户标识、登录时间、客户端信息、请求上下文与当前插件运行代际信息

#### Scenario: 登出成功事件触发 Hook
- **WHEN** 用户登出成功且宿主发布 `auth.logout.succeeded` Hook
- **THEN** 宿主向订阅该 Hook 的已启用插件分发事件
- **AND** 插件只能读取宿主公开的上下文字段

### Requirement: Hook 执行失败必须与主流程隔离
系统 SHALL 对插件 Hook 的超时、异常和返回错误实施隔离，不得让插件扩展破坏宿主主链路。

#### Scenario: 插件 Hook 执行失败
- **WHEN** 某插件在登录成功 Hook 中超时、崩溃或返回错误
- **THEN** 用户登录主流程仍然返回成功
- **AND** 宿主记录该插件的执行失败信息
- **AND** 其他插件的 Hook 仍按顺序继续执行或按策略被安全跳过

### Requirement: 宿主发布前端 Slot 扩展点
系统 SHALL 为前端页面和布局发布受控的 Slot 扩展点，允许插件在宿主公开位置插入 UI 内容。

#### Scenario: 插件向宿主布局插入内容
- **WHEN** 一个已启用插件声明向 `layout.user-dropdown.after` 插入前端内容
- **THEN** 宿主在该 Slot 对应位置尝试加载插件声明的前端入口
- **AND** 源码插件的 Slot 内容必须来自真实前端源码文件，而不是仅依赖声明式 JSON 配置
- **AND** 这些源码文件默认放在 `frontend/src/slots/` 目录下，并由宿主在构建时发现和挂载
- **AND** 插件内容只在宿主公开的容器范围内渲染
- **AND** 插件不能越权访问未公开的宿主内部实现

#### Scenario: 插件向右上角菜单栏插入页面入口
- **WHEN** 一个已启用插件声明向 `layout.user-dropdown.after` 插入插件菜单入口
- **THEN** 宿主在右上角菜单栏展示该入口文案
- **AND** 点击该入口后宿主以内页导航方式打开插件 Tab 页面
- **AND** 该过程不会触发整页刷新

#### Scenario: 登录态在线启用插件后立即激活右上角入口路由
- **WHEN** 管理员在当前已登录会话中启用一个会向 `layout.user-dropdown.after` 注入入口的插件
- **THEN** 宿主无需重新登录即可同步刷新该入口对应的动态路由
- **AND** 用户点击该入口后不会进入 404 页面
- **AND** 宿主直接以内页 Tab 方式打开插件页面

#### Scenario: 当前会话在插件状态变化后重新获得焦点
- **WHEN** 当前已登录会话之外的其他操作改变了会注入 `layout.user-dropdown.after` 的插件状态，且当前标签页重新获得焦点
- **THEN** 宿主自动同步该 Slot 的可见性与对应动态路由
- **AND** 已启用插件的右上角入口重新显示且可正常打开
- **AND** 已禁用插件的右上角入口及时隐藏

#### Scenario: 插件 Slot 契约不匹配
- **WHEN** 插件声明的前端入口与 Slot 所要求的契约不兼容
- **THEN** 宿主跳过该插件内容渲染
- **AND** 宿主记录契约不匹配错误
- **AND** 当前页面其他宿主内容正常渲染

### Requirement: Hook 与 Slot 执行顺序可预测
系统 SHALL 为同一 Hook 或 Slot 上的多个插件定义稳定的执行顺序。

#### Scenario: 多个插件订阅同一 Hook
- **WHEN** 多个插件同时订阅同一个后端 Hook 或前端 Slot
- **THEN** 宿主按照 manifest 显式优先级或统一的默认排序规则执行
- **AND** 相同输入下的执行顺序在各节点保持一致
