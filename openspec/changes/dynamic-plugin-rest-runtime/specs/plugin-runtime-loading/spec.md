## ADDED Requirements

### Requirement: 动态插件运行时产物携带可治理的路由合同

系统 SHALL 允许动态插件在运行时产物中携带后端动态路由合同，宿主装载产物后能够恢复这些路由的路径、方法与最小治理元数据，而不需要在请求时再次扫描源码目录。

#### Scenario: 构建阶段提取动态路由合同

- **WHEN** 构建动态插件运行时产物
- **THEN** 构建器从`backend/api/**/*.go`中的请求结构体`g.Meta`提取动态路由元数据
- **AND** 将这些元数据写入运行时产物中的专用区段
- **AND** 宿主加载产物后可恢复为动态插件`manifest.Routes`

#### Scenario: 宿主校验动态路由合同

- **WHEN** 宿主装载一个动态插件的路由合同
- **THEN** 宿主校验内部路径、方法、`access`、`permission`与`operLog`是否合法
- **AND** `access`未声明时按`login`处理
- **AND** `public`路由不得声明`permission`
- **AND** 非法合同会导致该产物装载失败

### Requirement: 宿主按固定前缀分发动态插件路由

系统 SHALL 将动态插件公开接口固定在`/api/v1/extensions/{pluginId}/...`下，并仅让命中该前缀的请求进入动态插件路由分发链路。

#### Scenario: 非插件请求不进入动态分发链路

- **WHEN** 宿主收到一个未命中`/api/v1/extensions/{pluginId}/...`的请求
- **THEN** 该请求继续走宿主原有路由链
- **AND** 不会触发动态插件路由匹配

#### Scenario: 宿主按`pluginId`与内部路径匹配动态路由

- **WHEN** 宿主收到一个命中固定前缀的请求
- **THEN** 宿主先提取`pluginId`
- **AND** 仅在该插件的已启用动态路由集合内按方法和内部路径做匹配
- **AND** 支持`/path/{id}`形式的动态路径段匹配

### Requirement: 动态路由通过受限`Wasm bridge`执行并保留占位回退

系统 SHALL 在完成路由命中与治理校验后，优先通过当前激活版本声明的受限`Wasm bridge`执行业务路由；若当前产物未声明可执行 bridge，则回退到明确的`501`占位响应。

#### Scenario: 命中受保护动态路由

- **WHEN** 一个登录型动态路由被成功匹配
- **THEN** 宿主先完成登录校验
- **AND** 若该路由声明了`permission`，宿主继续完成权限校验
- **AND** 若当前激活产物声明了可执行 bridge，则宿主通过`Wasm bridge`执行该路由并回写真实响应
- **AND** 若当前激活产物未声明可执行 bridge，则宿主返回`501`占位响应

#### Scenario: 命中公开动态路由

- **WHEN** 一个`public`动态路由被成功匹配
- **THEN** 宿主不得解析登录令牌
- **AND** 宿主不得注入用户业务上下文
- **AND** 若当前激活产物声明了可执行 bridge，则宿主通过`Wasm bridge`执行该路由并回写真实响应
- **AND** 若当前激活产物未声明可执行 bridge，则该路由返回`501`占位响应
