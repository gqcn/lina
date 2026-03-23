# API 接口规范

## 定时任务管理接口

### 1. 查询任务列表

**接口**: `GET /job/list`

**请求参数**:
```typescript
{
  name?: string        // 任务名称(模糊查询)
  group?: string       // 任务分组
  status?: number      // 状态: 1=启用 0=禁用
  page: number         // 页码
  pageSize: number     // 每页数量
}
```

**响应数据**:
```typescript
{
  items: [
    {
      id: number
      name: string
      group: string
      command: string
      cronExpr: string
      description: string
      status: number
      singleton: number
      maxTimes: number
      execTimes: number
      isSystem: number
      createBy: string
      createTime: string
      updateBy: string
      updateTime: string
      remark: string
    }
  ],
  total: number
}
```

### 2. 创建任务

**接口**: `POST /job/create`

**请求数据**:
```typescript
{
  name: string          // 任务名称
  group: string         // 任务分组
  command: string       // 执行指令
  cronExpr: string      // Cron表达式
  description?: string  // 任务描述
  status: number        // 状态: 1=启用 0=禁用
  singleton: number     // 执行模式: 1=单例 0=并行
  maxTimes: number      // 最大执行次数, 0表示无限制
  remark?: string       // 备注
}
```

**响应数据**: 标准响应

### 3. 更新任务

**接口**: `PUT /job/update`

**请求数据**:
```typescript
{
  id: number
  name: string
  group: string
  command: string       // 系统任务不可修改
  cronExpr: string
  description?: string
  status: number
  singleton: number
  maxTimes: number
  remark?: string
}
```

**响应数据**: 标准响应

### 4. 删除任务

**接口**: `DELETE /job/delete`

**请求数据**:
```typescript
{
  ids: number[]  // 任务ID列表
}
```

**响应数据**: 标准响应

**业务规则**: 系统任务(is_system=1)不可删除

### 5. 更新任务状态

**接口**: `PUT /job/status`

**请求数据**:
```typescript
{
  id: number
  status: number  // 1=启用 0=禁用
}
```

**响应数据**: 标准响应

### 6. 手动执行任务

**接口**: `POST /job/run`

**请求数据**:
```typescript
{
  id: number
}
```

**响应数据**: 标准响应

## 执行日志接口

### 7. 查询执行日志列表

**接口**: `GET /job/log/list`

**请求参数**:
```typescript
{
  jobName?: string      // 任务名称(模糊查询)
  status?: number       // 执行状态: 1=成功 0=失败
  startTime?: string    // 开始时间(起)
  endTime?: string      // 开始时间(止)
  page: number
  pageSize: number
}
```

**响应数据**:
```typescript
{
  items: [
    {
      id: number
      jobId: number
      jobName: string
      jobGroup: string
      command: string
      status: number
      startTime: string
      endTime: string
      duration: number
      errorMsg: string
      createTime: string
    }
  ],
  total: number
}
```
