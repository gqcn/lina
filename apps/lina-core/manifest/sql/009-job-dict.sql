-- 009: Job module dictionary definitions

-- ============================================================
-- 字典初始化数据：任务状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('任务状态', 'sys_job_status', 1, '定时任务状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_status', '启用', '1', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_status', '禁用', '0', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典初始化数据：执行模式
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('执行模式', 'sys_job_exec_mode', 1, '定时任务执行模式列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_exec_mode', '单例', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_exec_mode', '并行', '0', 2, 'default', 1, NOW(), NOW());

-- ============================================================
-- 字典初始化数据：系统任务标识
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('系统任务', 'sys_job_system_flag', 1, '定时任务系统标识列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_system_flag', '是', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_job_system_flag', '否', '0', 2, 'default', 1, NOW(), NOW());