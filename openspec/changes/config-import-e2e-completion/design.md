---
name: config-import-e2e-completion
description: 参数设置导入功能 E2E 测试完善的设计文档
type: project
---

# 设计文档

## 现有实现分析

### 后端 API

- `POST /config/import` - 导入参数设置，支持 `updateSupport` 参数
- `GET /config/import-template` - 下载 Excel 模板
- 响应格式：`{ success: number, fail: number, failList: [{row, reason}] }`

### 前端组件

- `config-import-modal.vue` - 导入弹窗
  - 文件拖拽上传
  - 覆盖模式开关
  - 下载模板链接
  - 导入结果展示（Modal.success/error）

### 现有 E2E 测试

`TC0055-config-import.ts` 覆盖：
- ✅ 打开导入弹窗
- ✅ 下载模板链接存在
- ✅ 拖拽上传区域和覆盖开关 UI
- ✅ 下载模板接口正确

## 缺失测试场景

1. **正常导入流程**：上传有效 Excel 文件，验证导入成功
2. **覆盖模式验证**：
   - 不开启覆盖：重复 key 应跳过并报告失败
   - 开启覆盖：重复 key 应更新已有记录
3. **导入结果验证**：
   - 成功提示消息
   - 部分失败提示消息
4. **数据正确性**：导入后验证数据库中数据正确
5. **无效文件处理**：上传非 Excel 文件

## 测试设计

### 测试数据准备

需要准备测试用 Excel 文件：
- `config-import-valid.xlsx` - 有效数据
- `config-import-duplicate.xlsx` - 包含重复 key

### Page Object 扩展

`ConfigPage.ts` 需要添加方法：
- `uploadImportFile(filePath)` - 上传导入文件
- `toggleUpdateSupport(enabled)` - 切换覆盖开关
- `confirmImport()` - 确认导入
- `getImportResult()` - 获取导入结果消息