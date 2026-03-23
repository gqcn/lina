-- v0.8.0 定时任务功能

-- 创建定时任务表
CREATE TABLE IF NOT EXISTS `sys_job` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `name` varchar(64) NOT NULL COMMENT '任务名称',
  `group` varchar(64) NOT NULL DEFAULT 'default' COMMENT '任务分组',
  `command` varchar(500) NOT NULL COMMENT '执行指令',
  `cron_expr` varchar(255) NOT NULL COMMENT 'Cron表达式',
  `description` varchar(500) DEFAULT NULL COMMENT '任务描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=禁用',
  `singleton` tinyint NOT NULL DEFAULT 1 COMMENT '执行模式：1=单例 0=并行',
  `max_times` int NOT NULL DEFAULT 0 COMMENT '最大执行次数，0表示无限制',
  `exec_times` int NOT NULL DEFAULT 0 COMMENT '已执行次数',
  `is_system` tinyint NOT NULL DEFAULT 0 COMMENT '是否系统任务：1=是 0=否',
  `create_by` varchar(64) DEFAULT NULL COMMENT '创建者',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_by` varchar(64) DEFAULT NULL COMMENT '更新者',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_group` (`group`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';

-- 创建任务执行日志表
CREATE TABLE IF NOT EXISTS `sys_job_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `job_id` bigint unsigned NOT NULL COMMENT '任务ID',
  `job_name` varchar(64) NOT NULL COMMENT '任务名称',
  `job_group` varchar(64) NOT NULL COMMENT '任务分组',
  `command` varchar(500) NOT NULL COMMENT '执行指令',
  `status` tinyint NOT NULL COMMENT '执行状态：1=成功 0=失败',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `duration` int DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `error_msg` text COMMENT '错误信息',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表';

-- 创建分布式锁表
CREATE TABLE IF NOT EXISTS `sys_locker` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '锁ID',
  `name` varchar(255) NOT NULL COMMENT '锁名称',
  `reason` varchar(500) DEFAULT NULL COMMENT '锁定原因',
  `create_time` datetime NOT NULL COMMENT '创建时间',
  `expire_time` datetime NOT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';

-- 插入系统任务
INSERT INTO `sys_job` VALUES
(1, '会话清理', 'system', '<session.Cleanup>', '0 0 * * * *', '清理过期的用户会话', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL),
(2, '服务器监控', 'system', '<servermon.Collect>', '0 * * * * *', '采集服务器性能指标', 1, 1, 0, 0, 1, 'admin', NOW(), NULL, NULL, NULL);

-- 插入字典类型：任务状态
INSERT INTO `sys_dict_type` (`name`, `dict_type`, `status`, `create_by`, `create_time`, `remark`)
VALUES ('任务状态', 'sys_job_status', 1, 'admin', NOW(), '定时任务状态');

-- 插入字典数据：任务状态
INSERT INTO `sys_dict_data` (`dict_type_id`, `label`, `value`, `sort`, `tag_type`, `create_by`, `create_time`)
SELECT id, '启用', '1', 1, 'success', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_status'
UNION ALL
SELECT id, '禁用', '0', 2, 'danger', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_status';

-- 插入字典类型：任务执行模式
INSERT INTO `sys_dict_type` (`name`, `dict_type`, `status`, `create_by`, `create_time`, `remark`)
VALUES ('任务执行模式', 'sys_job_singleton', 1, 'admin', NOW(), '定时任务执行模式');

-- 插入字典数据：任务执行模式
INSERT INTO `sys_dict_data` (`dict_type_id`, `label`, `value`, `sort`, `tag_type`, `create_by`, `create_time`)
SELECT id, '单例执行', '1', 1, 'default', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_singleton'
UNION ALL
SELECT id, '并行执行', '0', 2, 'default', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_singleton';

-- 插入字典类型：任务执行状态
INSERT INTO `sys_dict_type` (`name`, `dict_type`, `status`, `create_by`, `create_time`, `remark`)
VALUES ('任务执行状态', 'sys_job_log_status', 1, 'admin', NOW(), '任务执行日志状态');

-- 插入字典数据：任务执行状态
INSERT INTO `sys_dict_data` (`dict_type_id`, `label`, `value`, `sort`, `tag_type`, `create_by`, `create_time`)
SELECT id, '成功', '1', 1, 'success', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_log_status'
UNION ALL
SELECT id, '失败', '0', 2, 'danger', 'admin', NOW() FROM `sys_dict_type` WHERE `dict_type` = 'sys_job_log_status';
