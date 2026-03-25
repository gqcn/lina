# Dictionary Export/Import Capability

## Overview

字典管理导出导入功能，支持同时处理字典类型和字典数据，简化用户操作流程。

---

### Requirement: Export File Naming

导出文件名称 SHALL 使用描述性命名，格式为 `{功能描述}导出.xlsx`。

#### Scenario: Export dictionary types
WHEN 用户导出字典类型数据
THEN 导出文件名 SHALL 为 `字典类型导出.xlsx`

#### Scenario: Export dictionary data
WHEN 用户导出字典数据
THEN 导出文件名 SHALL 为 `字典数据导出.xlsx`

#### Scenario: Export config data
WHEN 用户导出参数设置数据
THEN 导出文件名 SHALL 为 `参数设置导出.xlsx`

---

### Requirement: Combined Dictionary Export

字典类型面板的导出功能 SHALL 同时导出字典类型和字典数据。

#### Scenario: Export all dictionary data
WHEN 用户在字典类型面板点击导出按钮且未选中任何记录
THEN 系统 SHALL 导出所有字典类型和所有字典数据
AND Excel 文件包含两个 Sheet：字典类型、字典数据
AND 导出文件名 SHALL 为 `字典管理导出.xlsx`

#### Scenario: Export selected dictionary types
WHEN 用户选中若干字典类型后点击导出按钮
THEN 系统 SHALL 导出选中的字典类型及其关联的字典数据
AND Sheet 1 包含选中的字典类型
AND Sheet 2 包含这些类型下的所有字典数据

---

### Requirement: Combined Dictionary Import

字典类型面板的导入功能 SHALL 支持同时导入字典类型和字典数据。

#### Scenario: Import dictionary with both sheets
WHEN 用户上传包含字典类型和字典数据两个 Sheet 的 Excel 文件
THEN 系统 SHALL 先导入字典类型，再导入字典数据
AND 返回类型和数据各自的导入结果

#### Scenario: Import with duplicate type
WHEN 导入的字典类型已存在（type 字段重复）
THEN 系统 SHALL 跳过该记录
AND 在失败列表中记录失败原因"字典类型已存在"

#### Scenario: Import data with missing type reference
WHEN 导入的字典数据引用了不存在的字典类型（不在 Sheet 1 也不在数据库）
THEN 系统 SHALL 跳过该记录
AND 在失败列表中记录失败原因"字典类型不存在"

---

### Requirement: Remove Data Panel Export/Import

字典数据面板 SHALL 移除独立的导出和导入按钮。

#### Scenario: Data panel toolbar
WHEN 用户查看字典数据面板
THEN 工具栏 SHALL 不显示导出按钮
AND 工具栏 SHALL 不显示导入按钮
AND 工具栏 SHALL 保留新增和删除按钮

---

### Requirement: Export/Import Template

系统 SHALL 提供合并导入模板下载功能。

#### Scenario: Download import template
WHEN 用户点击下载模板按钮
THEN 系统 SHALL 返回包含两个 Sheet 的 Excel 模板
AND Sheet 1 包含字典类型的示例数据行和字段说明
AND Sheet 2 包含字典数据的示例数据行和字段说明
AND 文件名 SHALL 为 `字典管理导入模板.xlsx`