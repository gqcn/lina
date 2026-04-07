# plugin-demo backend

该目录存放 `plugin-demo` 的**后端 Go 源码实现**，宿主只提供通用注册表与执行器：

- `plugin.go`：在编译期注册插件订阅的宿主 Hook 与后端资源
- `plugin.yaml` 不负责注册后端路由；源码插件的后端能力以本目录 Go 注册为准

## 当前示例

- `plugin.go`
  - 注册 `auth.login.succeeded` Hook
  - 注册 `login-audits` 资源
  - 由宿主通用执行器继续装配为可分页查询的数据源

## 设计边界

- `plugin-demo` 的业务语义、表结构、资源声明都维护在本目录或插件自身 `manifest/` 下
- 宿主 `lina-core` 只维护**插件框架能力**与通用资源查询入口，不再手写 `plugin-demo` 专属控制器、服务或路由逻辑
