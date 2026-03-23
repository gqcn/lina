# 前端规范

## 页面结构

### 1. 任务列表页 (apps/lina-vben/apps/web-antd/src/views/system/job/index.vue)

#### 搜索区域
- 任务名称: Input (模糊查询)
- 任务分组: Input
- 任务状态: Select (字典: sys_job_status)
- 搜索按钮、重置按钮

#### 工具栏
- 新增按钮: 打开任务表单弹窗
- 刷新按钮: 重新加载列表

#### 表格 (VXE-Grid)
| 列名 | 字段 | 说明 |
|------|------|------|
| 任务名称 | name | - |
| 任务分组 | group | - |
| 执行指令 | command | 系统任务显示 `<函数名>` |
| Cron表达式 | cronExpr | - |
| 状态 | status | 字典标签 (sys_job_status) |
| 执行模式 | singleton | 字典标签 (sys_job_singleton) |
| 执行次数 | execTimes/maxTimes | 显示为 "已执行/最大次数" |
| 操作 | - | 编辑、删除、启用/禁用、执行、日志 |

#### 操作列按钮
- **编辑**: 打开编辑表单弹窗
- **删除**: 确认后删除 (系统任务不显示)
- **启用/禁用**: 切换任务状态
- **执行一次**: 手动触发任务执行
- **查看日志**: 跳转到日志页面并筛选当前任务

#### 表单弹窗 (Modal)
**字段**:
- 任务名称: Input (必填)
- 任务分组: Input (必填)
- 执行指令: Input (必填, 系统任务只读)
- Cron表达式: Input (必填)
- 任务描述: Textarea
- 任务状态: RadioGroup (字典: sys_job_status)
- 执行模式: RadioGroup (字典: sys_job_singleton)
- 最大执行次数: InputNumber (0表示无限制)
- 备注: Textarea

**校验规则**:
- 任务名称: 必填, 最大64字符
- 任务分组: 必填, 最大64字符
- 执行指令: 必填, 最大500字符
- Cron表达式: 必填, 格式校验

### 1.1. 任务表单组件 (apps/lina-vben/apps/web-antd/src/views/system/job/form.vue)

独立的表单组件,用于新增和编辑任务。

**Props**:
- `formApi`: VbenFormApi 实例

**字段定义**: 与表单弹窗字段相同

### 2. 执行日志页 (apps/lina-vben/apps/web-antd/src/views/system/job/log.vue)

#### 搜索区域
- 任务名称: Input (模糊查询)
- 执行状态: Select (字典: sys_job_log_status)
- 开始时间: RangePicker
- 搜索按钮、重置按钮

#### 表格 (VXE-Grid)
| 列名 | 字段 | 说明 |
|------|------|------|
| 任务名称 | jobName | - |
| 任务分组 | jobGroup | - |
| 执行指令 | command | - |
| 开始时间 | startTime | 格式: YYYY-MM-DD HH:mm:ss |
| 结束时间 | endTime | 格式: YYYY-MM-DD HH:mm:ss |
| 执行耗时 | duration | 显示为 "XXXms" |
| 执行状态 | status | 字典标签 (sys_job_log_status) |
| 操作 | - | 查看详情 |

#### 详情抽屉 (Drawer)
显示完整的执行日志信息:
- 任务名称
- 任务分组
- 执行指令
- 开始时间
- 结束时间
- 执行耗时
- 执行状态
- 错误信息 (如果失败)

## MODIFIED Requirements

### Requirement: 定时任务菜单位置
定时任务功能必须放置在"系统管理"菜单下,而非"系统监控"菜单下。

#### Scenario: 菜单导航
WHEN 用户登录系统后查看左侧菜单
THEN 应在"系统管理"菜单下看到"定时任务"菜单项
AND 点击后可正常访问定时任务列表页

## ADDED Requirements

### Requirement: 任务表单组件
任务列表页必须包含独立的表单组件文件,用于新增和编辑任务。

#### Scenario: 表单组件加载
WHEN 用户点击"新增"或"编辑"按钮
THEN 系统应成功加载 JobForm 组件并显示表单弹窗
AND 不应出现组件加载错误

## 路由配置

```typescript
// apps/lina-vben/apps/web-antd/src/router/routes/modules/system.ts
{
  path: 'job',
  name: 'Job',
  component: () => import('#/views/system/job/index.vue'),
  meta: {
    icon: 'lucide:clock',
    title: '定时任务',
  },
},
{
  path: 'job/log',
  name: 'JobLog',
  component: () => import('#/views/system/job/log.vue'),
  meta: {
    hideInMenu: true,
    title: '执行日志',
  },
}
```

## API 定义

```typescript
// apps/lina-vben/apps/web-antd/src/api/system/job.ts
import { requestClient } from '#/api/request';

export namespace JobApi {
  export interface Job {
    id: number;
    name: string;
    group: string;
    command: string;
    cronExpr: string;
    description?: string;
    status: number;
    singleton: number;
    maxTimes: number;
    execTimes: number;
    isSystem: number;
    createBy?: string;
    createTime?: string;
    updateBy?: string;
    updateTime?: string;
    remark?: string;
  }

  export interface JobLog {
    id: number;
    jobId: number;
    jobName: string;
    jobGroup: string;
    command: string;
    status: number;
    startTime: string;
    endTime?: string;
    duration?: number;
    errorMsg?: string;
    createTime?: string;
  }

  export interface ListParams {
    name?: string;
    group?: string;
    status?: number;
    page: number;
    pageSize: number;
  }

  export interface LogListParams {
    jobName?: string;
    status?: number;
    startTime?: string;
    endTime?: string;
    page: number;
    pageSize: number;
  }
}

export const jobApi = {
  list: (params: JobApi.ListParams) =>
    requestClient.get<{ items: JobApi.Job[]; total: number }>('/job/list', { params }),

  create: (data: Partial<JobApi.Job>) =>
    requestClient.post('/job/create', data),

  update: (data: Partial<JobApi.Job>) =>
    requestClient.put('/job/update', data),

  delete: (ids: number[]) =>
    requestClient.delete('/job/delete', { data: { ids } }),

  updateStatus: (id: number, status: number) =>
    requestClient.put('/job/status', { id, status }),

  run: (id: number) =>
    requestClient.post('/job/run', { id }),

  logList: (params: JobApi.LogListParams) =>
    requestClient.get<{ items: JobApi.JobLog[]; total: number }>('/job/log/list', { params }),
};
```

## 组件使用规范

### 表单组件
- 使用 `useVbenForm` 创建表单
- RadioGroup 使用 `optionType: 'button'` + `buttonStyle: 'solid'`
- 字典数据使用 `useDictStore` 获取

### 表格组件
- 使用 `useVbenVxeGrid` + `Page` 组件
- 操作列使用 `ghost-button` + `Popconfirm`
- 状态列使用 `DictTag` 组件显示字典标签

### 弹窗组件
- 使用 `useVbenModal` 创建弹窗
- 使用 `useVbenDrawer` 创建抽屉

## 交互规范

1. **新增任务**: 点击新增按钮 → 打开表单弹窗 → 填写信息 → 提交 → 刷新列表
2. **编辑任务**: 点击编辑按钮 → 打开表单弹窗(回显数据) → 修改信息 → 提交 → 刷新列表
3. **删除任务**: 点击删除按钮 → 二次确认 → 删除 → 刷新列表
4. **启用/禁用**: 点击按钮 → 直接切换状态 → 刷新列表
5. **执行任务**: 点击执行按钮 → 二次确认 → 执行 → 提示结果
6. **查看日志**: 点击日志按钮 → 跳转到日志页面并筛选当前任务
7. **查看日志详情**: 点击详情按钮 → 打开详情抽屉 → 显示完整信息
