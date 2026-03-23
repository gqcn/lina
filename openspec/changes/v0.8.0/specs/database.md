# 数据库规范

## 表结构设计

### sys_job (定时任务表)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint unsigned | PK, AUTO_INCREMENT | 任务ID |
| name | varchar(64) | NOT NULL | 任务名称 |
| group | varchar(64) | NOT NULL, DEFAULT 'default' | 任务分组 |
| command | varchar(500) | NOT NULL | 执行指令 |
| cron_expr | varchar(255) | NOT NULL | Cron表达式 |
| description | varchar(500) | NULL | 任务描述 |
| status | tinyint | NOT NULL, DEFAULT 1 | 状态: 1=启用 0=禁用 |
| singleton | tinyint | NOT NULL, DEFAULT 1 | 执行模式: 1=单例 0=并行 |
| max_times | int | NOT NULL, DEFAULT 0 | 最大执行次数, 0=无限制 |
| exec_times | int | NOT NULL, DEFAULT 0 | 已执行次数 |
| is_system | tinyint | NOT NULL, DEFAULT 0 | 是否系统任务: 1=是 0=否 |
| create_by | varchar(64) | NULL | 创建者 |
| create_time | datetime | NULL | 创建时间 |
| update_by | varchar(64) | NULL | 更新者 |
| update_time | datetime | NULL | 更新时间 |
| remark | varchar(500) | NULL | 备注 |

**索引**:
- PRIMARY KEY: `id`
- INDEX: `idx_group` (`group`)
- INDEX: `idx_status` (`status`)

### sys_job_log (任务执行日志表)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint unsigned | PK, AUTO_INCREMENT | 日志ID |
| job_id | bigint unsigned | NOT NULL | 任务ID |
| job_name | varchar(64) | NOT NULL | 任务名称 |
| job_group | varchar(64) | NOT NULL | 任务分组 |
| command | varchar(500) | NOT NULL | 执行指令 |
| status | tinyint | NOT NULL | 执行状态: 1=成功 0=失败 |
| start_time | datetime | NOT NULL | 开始时间 |
| end_time | datetime | NULL | 结束时间 |
| duration | int | NULL | 执行耗时(毫秒) |
| error_msg | text | NULL | 错误信息 |
| create_time | datetime | NULL | 创建时间 |

**索引**:
- PRIMARY KEY: `id`
- INDEX: `idx_job_id` (`job_id`)
- INDEX: `idx_status` (`status`)
- INDEX: `idx_start_time` (`start_time`)

### sys_locker (分布式锁表)

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint unsigned | PK, AUTO_INCREMENT | 锁ID |
| name | varchar(255) | NOT NULL, UNIQUE | 锁名称 |
| reason | varchar(500) | NULL | 锁定原因 |
| create_time | datetime | NOT NULL | 创建时间 |
| expire_time | datetime | NOT NULL | 过期时间 |

**索引**:
- PRIMARY KEY: `id`
- UNIQUE KEY: `uk_name` (`name`)
- INDEX: `idx_expire_time` (`expire_time`)

## 初始化数据

### 系统任务
```sql
INSERT INTO `sys_job` VALUES
(1, '会话清理', 'system', '<session.Cleanup>', '0 0 * * * *', '清理过期的用户会话', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL),
(2, '服务器监控', 'system', '<servermon.Collect>', '0 * * * * *', '采集服务器性能指标', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL);
```

### 字典类型和数据
```sql
-- 任务状态
INSERT INTO `sys_dict_type` VALUES (NULL, '任务状态', 'sys_job_status', 1, 'admin', NOW(), NULL, NULL, '定时任务状态');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_status'), '启用', '1', 1, 'success', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_status'), '禁用', '0', 2, 'danger', 'admin', NOW(), NULL, NULL, NULL);

-- 执行模式
INSERT INTO `sys_dict_type` VALUES (NULL, '任务执行模式', 'sys_job_singleton', 1, 'admin', NOW(), NULL, NULL, '定时任务执行模式');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_singleton'), '单例执行', '1', 1, 'default', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_singleton'), '并行执行', '0', 2, 'default', 'admin', NOW(), NULL, NULL, NULL);

-- 执行状态
INSERT INTO `sys_dict_type` VALUES (NULL, '任务执行状态', 'sys_job_log_status', 1, 'admin', NOW(), NULL, NULL, '任务执行日志状态');
INSERT INTO `sys_dict_data` VALUES
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_log_status'), '成功', '1', 1, 'success', 'admin', NOW(), NULL, NULL, NULL),
(NULL, (SELECT id FROM sys_dict_type WHERE dict_type='sys_job_log_status'), '失败', '0', 2, 'danger', 'admin', NOW(), NULL, NULL, NULL);
```
