## 1. 数据库建表与代码生成

- [x] 1.1 创建 `manifest/sql/v0.2.0.sql`，包含 sys_dict_type、sys_dict_data、sys_dept、sys_post、sys_user_dept、sys_user_post 六张表的建表语句和初始化数据（字典、部门、岗位种子数据）
- [x] 1.2 执行 `make dao` 生成 DAO/DO/Entity 代码

## 2. 字典管理后端

- [x] 2.1 创建字典类型 API 定义 `api/dict/v1/dict_type.go`（列表、新增、修改、删除、详情、导出、选项列表）
- [x] 2.2 创建字典数据 API 定义 `api/dict/v1/dict_data.go`（列表、新增、修改、删除、详情、导出、按类型获取选项）
- [x] 2.3 执行 `make ctrl` 生成字典控制器骨架
- [x] 2.4 实现字典类型 Service（`internal/service/dict/dict_type.go`）：列表查询（分页+筛选）、CRUD、导出、选项列表、类型唯一性校验、删除时检查关联数据
- [x] 2.5 实现字典数据 Service（`internal/service/dict/dict_data.go`）：按类型查询列表（分页+筛选）、CRUD、导出、按类型获取选项（排序+缓存用）
- [x] 2.6 填写字典类型 Controller 业务逻辑
- [x] 2.7 填写字典数据 Controller 业务逻辑
- [x] 2.8 在 `internal/cmd/cmd.go` 中注册字典管理路由

## 3. 字典管理前端

- [x] 3.1 创建字典 API 文件 `src/api/system/dict/dict-type.ts`、`dict-type-model.d.ts`、`dict-data.ts`、`dict-data-model.d.ts`
- [x] 3.2 实现 Pinia 字典缓存 Store（`src/store/dict.ts`）：缓存管理、请求去重、刷新缓存
- [x] 3.3 实现全局 DictTag 组件（`src/components/dict/`）：Tag 样式预设数据（12 种颜色）、DictTag 渲染组件（预设色/自定义色/CSS 类/加载中/fallback）
- [x] 3.4 实现 TagStylePicker 组件（`src/views/system/dict/data/tag-style-picker.vue`）：默认色/自定义色模式切换、预设色下拉（带颜色预览）、hex 输入框
- [x] 3.5 实现字典管理双面板主页面（`src/views/system/dict/index.vue`）和 Mitt 事件通信（`mitt.ts`）
- [x] 3.6 实现字典类型列表页（`src/views/system/dict/type/index.vue`、`data.ts`）：VXE-Grid 列表、搜索表单、新增/删除/导出/刷新缓存工具栏、行点击事件
- [x] 3.7 实现字典类型编辑弹窗（`dict-type-modal.vue`）：Modal 表单、字段校验、变更检测
- [x] 3.8 实现字典数据列表页（`src/views/system/dict/data/index.vue`、`data.ts`）：VXE-Grid 列表、标签列 Tag 渲染、搜索表单、新增/删除/导出工具栏
- [x] 3.9 实现字典数据编辑抽屉（`dict-data-drawer.vue`）：600px Drawer、2 列网格表单、TagStylePicker 集成
- [x] 3.10 在 `src/router/routes/modules/system.ts` 中添加字典管理路由

## 4. 部门管理后端

- [x] 4.1 创建部门 API 定义 `api/dept/v1/dept.go`（列表、新增、修改、删除、详情、树形结构、排除节点）
- [x] 4.2 执行 `make ctrl` 生成部门控制器骨架
- [x] 4.3 实现部门 Service（`internal/service/dept/dept.go`）：树形列表查询（筛选+排序）、CRUD（ancestors 自动计算、更新时同步子部门 ancestors）、删除校验（子部门/关联用户）、树形结构接口、排除节点接口、部门用户列表（通过 sys_user_dept）
- [x] 4.4 填写部门 Controller 业务逻辑
- [x] 4.5 在 `internal/cmd/cmd.go` 中注册部门管理路由

## 5. 部门管理前端

- [x] 5.1 创建部门 API 文件 `src/api/system/dept/index.ts`、`model.d.ts`
- [x] 5.2 实现 DeptTree 可复用组件（`src/views/system/user/dept-tree.vue`）：树形展示、搜索、刷新、v-model 绑定
- [x] 5.3 实现部门管理列表页（`src/views/system/dept/index.vue`、`data.ts`）：VXE-Grid 树形模式、展开/折叠按钮、双击切换、DictTag 状态渲染、行操作（编辑/新增子部门/删除）
- [x] 5.4 实现部门编辑抽屉（`dept-drawer.vue`）：600px Drawer、TreeSelect 上级部门（显示完整路径、排除自身及子部门）、负责人下拉（新增 disabled、编辑从部门用户中选）、电话/邮箱校验、状态 RadioGroup
- [x] 5.5 在 `src/router/routes/modules/system.ts` 中添加部门管理路由

## 6. 岗位管理后端

- [x] 6.1 创建岗位 API 定义 `api/post/v1/post.go`（列表、新增、修改、删除、详情、导出、部门树、按部门获取选项）
- [x] 6.2 执行 `make ctrl` 生成岗位控制器骨架
- [x] 6.3 实现岗位 Service（`internal/service/post/post.go`）：分页列表查询（按部门过滤+筛选）、CRUD（编码唯一性校验）、批量删除、删除校验（关联用户）、导出、部门树接口、按部门获取选项
- [x] 6.4 填写岗位 Controller 业务逻辑
- [x] 6.5 在 `internal/cmd/cmd.go` 中注册岗位管理路由

## 7. 岗位管理前端

- [x] 7.1 创建岗位 API 文件 `src/api/system/post/index.ts`、`model.d.ts`
- [x] 7.2 实现岗位管理列表页（`src/views/system/post/index.vue`、`data.ts`）：左侧 DeptTree（260px）+ 右侧 VXE-Grid 列表、部门筛选联动、多选框、批量删除、导出、DictTag 状态渲染
- [x] 7.3 实现岗位编辑抽屉（`post-drawer.vue`）：600px Drawer、2 列网格表单、TreeSelect 部门选择（完整路径）、状态 RadioGroup、备注 Textarea 全宽
- [x] 7.4 在 `src/router/routes/modules/system.ts` 中添加岗位管理路由

## 8. 用户管理扩展后端

- [x] 8.1 扩展用户 API 定义：CreateReq/UpdateReq 增加 deptId 和 postIds 字段、GetOneRes 增加 deptId/deptName/postIds 字段、新增 DeptTreeReq/DeptTreeRes
- [x] 8.2 执行 `make ctrl` 更新用户控制器
- [x] 8.3 扩展用户 Service：创建/更新时处理 sys_user_dept 和 sys_user_post 关联表（事务内先删后插）、列表查询 LEFT JOIN 返回 deptName、详情返回 deptId/postIds、实现 DeptTree 接口
- [x] 8.4 填写用户 DeptTree Controller 逻辑
- [x] 8.5 更新用户删除逻辑：删除用户时同步清理 sys_user_dept 和 sys_user_post 关联记录

## 9. 用户管理扩展前端

- [x] 9.1 更新用户 API 类型定义：User 接口增加 deptId/deptName/postIds 字段、新增 getDeptTree 方法
- [x] 9.2 修改用户管理列表页（`index.vue`）：增加左侧 DeptTree 筛选组件、部门选择时传入 deptId 过滤参数、增加部门名称列
- [x] 9.3 修改用户编辑抽屉（`user-drawer.vue`、`data.ts`）：增加部门 TreeSelect 字段（必填、完整路径、搜索）、增加岗位 Select 字段（多选、按部门联动加载、切换部门时清空）

## 10. E2E 测试

- [x] 10.1 `TC0012-dict-type-crud.ts`：字典类型的新增、编辑、删除测试
- [x] 10.2 `TC0013-dict-data-crud.ts`：字典数据的新增（含 Tag 样式选择）、编辑、删除测试
- [x] 10.3 `TC0014-dict-export.ts`：字典类型和字典数据的导出测试
- [x] 10.4 `TC0015-dept-crud.ts`：部门的新增根部门、新增子部门、编辑、删除测试，树形展开/折叠
- [x] 10.5 `TC0016-post-crud.ts`：岗位的新增、编辑、删除、批量删除测试
- [x] 10.6 `TC0017-post-dept-filter.ts`：岗位管理左侧部门树筛选测试
- [x] 10.7 `TC0018-post-export.ts`：岗位导出测试
- [x] 10.8 `TC0019-user-dept-filter.ts`：用户管理左侧部门树筛选测试
- [x] 10.9 `TC0020-user-dept-post-form.ts`：用户编辑表单中部门选择和岗位联动选择测试
