-- 008: Login Status and File Scene Dictionaries

-- ============================================================
-- 字典初始化数据：登录状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('登录状态', 'sys_login_status', 1, '登录日志状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_login_status', '成功', '0', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_login_status', '失败', '1', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典初始化数据：文件业务场景
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('文件业务场景', 'sys_file_scene', 1, '文件管理业务场景列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '用户头像', 'avatar', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告图片', 'notice_image', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告附件', 'notice_attachment', 3, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '其他', 'other', 4, 'default', 1, NOW(), NOW());