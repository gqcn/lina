## Feedback

- [x] **FB-1**：新增 `sys_file` 数据表及版本 SQL 文件 `v0.6.0.sql`，记录所有上传文件的元信息（文件名、原始名、大小、后缀、存储路径、上传者等）
- [x] **FB-2**：设计并实现后端文件存储抽象层（Storage 接口 + 本地存储实现），预留 OSS 扩展能力
- [x] **FB-3**：实现文件管理后端 API（`POST /file/upload` 上传、`GET /file` 列表、`GET /file/download/{id}` 下载、`DELETE /file/{ids}` 删除），生成 DAO/Controller 骨架
- [x] **FB-4**：创建前端通用文件上传 API 和 FileUpload / ImageUpload 组件，参考 ruoyi-plus-vben5 的上传组件设计
- [x] **FB-5**：新增文件管理页面（系统管理 > 文件管理），包含文件列表、搜索、上传弹窗、下载、批量删除、图片预览等功能，参考 ruoyi-plus-vben5 的 OSS 文件管理页面
- [x] **FB-6**：改造 TiptapEditor 富文本编辑器的图片上传，从 Base64 内嵌改为调用通用文件上传接口
- [x] **FB-7**：改造用户头像上传，使用通用文件上传接口替代原有独立实现（移除旧的 avatar 上传端点和静态文件服务路由）
- [x] **FB-8**：编写文件管理模块及改造功能的 E2E 测试用例
