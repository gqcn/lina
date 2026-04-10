# plugin-demo-dynamic

`plugin-demo-dynamic`是当前`plugin-framework`迭代提供的独立动态`wasm`样例插件。

与[`plugin-demo-source`](../plugin-demo-source/README.md)不同，这个插件不会以源码形式编译进宿主，而是用于演示当前动态插件契约下的一条最小闭环：

- 通过`plugin.yaml`中的`menus`元数据向宿主声明 1 个左侧菜单入口；
- 菜单在宿主主内容区打开一页简要说明；
- 页面提供 1 个“打开独立页面”按钮；
- 按钮会打开一个不依赖`Vben`前端框架的独立静态页面。

## 目录结构

```text
plugin-demo-dynamic/
  go.mod
  README.md
  plugin.yaml
  backend/
    plugin.go
    api/
    internal/controller/
    service/
  frontend/
    pages/
      mount.js
      standalone.html
  manifest/
    sql/
      # 当前样例未提供业务 SQL；若后续新增业务表迁移，可继续按目录约定补充
  temp/
    # 按需生成且被 Git 忽略：
    # plugin-demo-dynamic.wasm
```

## 动态行为

当插件完成安装并启用后，宿主会呈现以下行为：

- 左侧菜单显示 1 个名为`动态插件示例`的入口；
- 打开菜单后，宿主通过动态页面外壳和内嵌挂载`ESM`契约加载插件页面；
- 页面会展示：
  - 1 个标题；
  - 1 段简短说明；
  - 1 个名为`打开独立页面`的按钮；
- 点击按钮后，会在浏览器新标签页中打开`standalone.html`；
- `standalone.html`是纯静态页面，刻意不依赖`Vben`。

## 单一真相源

当前样例的单一真相源就是插件目录内的明文源码本身：

- `backend/`保存 1 份仅供 review 的后端示例代码；
- `frontend/pages/`保存宿主内嵌挂载入口和独立静态页；
- `plugin.yaml`保存插件基础信息和菜单元数据；
- `manifest/sql/`仅在需要业务迁移时保存安装与卸载`SQL`；
- `temp/`仅保存本地构建产物，不进入版本库。

动态元数据不再通过额外的`JSON`边车文件维护。执行`make wasm`时，构建器会基于当前源码树自动生成：

- `lina.plugin.dynamic`；
- 前端静态资源数量摘要；
- `SQL`资源数量摘要。

生成产物会被`Git`忽略：

- `temp/plugin-demo-dynamic.wasm`按需生成；
- `temp/`目录不应提交到仓库。

动态真正读取的也不是本地`temp/`目录。宿主会从`plugin.dynamic.storagePath`下上传或手工拷贝进去的`.wasm`文件中解析前端资源，并在内存中构建只读资源 bundle；宿主重启后，会在启动预热或请求时懒加载阶段重新构建这些 bundle。

## 构建流程

构建全部动态`wasm`样例插件：

```bash
make wasm
```

仅构建当前插件：

```bash
make wasm p=plugin-demo-dynamic
```

根级`make dev`和`make build`流程都会在启动或打包宿主前先执行同一套`make wasm`步骤，因此仓库本身不再需要提交编译后的`wasm`二进制文件。

通用构建入口由根目录下的`hack/build-wasm/`独立工具负责。该工具拥有自己的`go.mod`，并且不依赖`lina-core`宿主模块。

## 后端示例边界

`backend/`目录的存在，是为了让这个动态样例在 review 形态上与`plugin-demo-source`保持一致。

当前边界如下：

- Lina 动态插件已经支持把宿主定义的声明式后端契约嵌入到`wasm`包中；
- 当前这类契约主要覆盖动态`Hook`执行与通用资源声明，宿主会为其提供超时控制、错误隔离与卸载清理；
- Lina 动态插件当前仍然**不支持**动态执行插件自带的`Go`后端源码；
- 因此，`backend/plugin.go`仍然只是 review 示例，不会自动向宿主注册；
- 如果动态插件今天确实需要可执行的后端能力，应通过`backend/hooks/`或`backend/resources/`下的声明式文件接入，这些内容会在`make wasm`时被编译进自定义节。

## Review 关注点

人工 review 这个样例时，建议重点确认：

- `plugin.yaml`是否清晰标识该插件属于独立动态插件；
- `frontend/pages/mount.js`是否只依赖文档已发布的宿主`ESM`契约；
- `frontend/pages/standalone.html`是否保持框架无关；
- `plugin.yaml`里的`menus`是否只声明 1 个属于该插件的左侧菜单；
- 执行`make wasm p=plugin-demo-dynamic`后，是否会生成`temp/plugin-demo-dynamic.wasm`；
- 动态契约测试是否仍能证明生成出的`wasm`与明文源码树保持一致；
- 未来新增的`backend/hooks/*.yaml`或`backend/resources/*.yaml`是否仍严格遵守已发布的声明式动态`ABI`，而不是假设宿主会执行任意`Go`代码。
