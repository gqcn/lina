-- 008: Menu Management, Role Management, User-Role Association

-- ============================================================
-- 菜单表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_menu (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '菜单ID',
    parent_id   INT          NOT NULL DEFAULT 0  COMMENT '父菜单ID（0=根菜单）',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '菜单名称（支持i18n）',
    path        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '路由地址',
    component   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '组件路径',
    perms       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '权限标识',
    icon        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '菜单图标',
    type        CHAR(1)      NOT NULL DEFAULT 'M' COMMENT '菜单类型（D=目录 M=菜单 B=按钮）',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    visible     TINYINT      NOT NULL DEFAULT 1  COMMENT '是否显示（1=显示 0=隐藏）',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0=停用 1=正常）',
    is_frame    TINYINT      NOT NULL DEFAULT 0  COMMENT '是否外链（1=是 0=否）',
    is_cache    TINYINT      NOT NULL DEFAULT 0  COMMENT '是否缓存（1=是 0=否）',
    query_param VARCHAR(255) NOT NULL DEFAULT '' COMMENT '路由参数（JSON格式）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='菜单权限表';

-- ============================================================
-- 角色表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '角色ID',
    name        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色名称',
    `key`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '权限字符',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    data_scope  TINYINT      NOT NULL DEFAULT 1  COMMENT '数据权限范围（1=全部 2=本部门 3=仅本人）',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0=停用 1=正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE(`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色信息表';

-- ============================================================
-- 角色-菜单关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role_menu (
    role_id INT NOT NULL COMMENT '角色ID',
    menu_id INT NOT NULL COMMENT '菜单ID',
    PRIMARY KEY (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色与菜单关联表';

-- ============================================================
-- 用户-角色关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_role (
    user_id INT NOT NULL COMMENT '用户ID',
    role_id INT NOT NULL COMMENT '角色ID',
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与角色关联表';

-- ============================================================
-- 字典类型: 菜单状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单状态', 'sys_menu_status', 1, '菜单状态列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 显示状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('显示状态', 'sys_show_hide', 1, '显示状态列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 菜单类型
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单类型', 'sys_menu_type', 1, '菜单类型列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 数据权限范围
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('数据权限范围', 'sys_data_scope', 1, '数据权限范围列表', NOW(), NOW());

-- ============================================================
-- 字典数据: 菜单状态
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '正常', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '停用', '0', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 显示状态
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '显示', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '隐藏', '0', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 菜单类型
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '目录', 'D', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '菜单', 'M', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '按钮', 'B', 3, 'warning', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 数据权限范围
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '全部数据', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '本部门数据', '2', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '仅本人数据', '3', 3, 'warning', 1, NOW(), NOW());

-- ============================================================
-- 初始化角色数据
-- ============================================================
INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('超级管理员', 'admin', 1, 1, 1, '超级管理员，拥有所有权限', NOW(), NOW());
INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('普通用户', 'user', 2, 3, 1, '普通用户，仅查看本人数据', NOW(), NOW());

-- ============================================================
-- 初始化菜单数据
-- ============================================================

-- 系统管理（目录）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1, 0, '系统管理', 'system', '', '', 'ant-design:setting-outlined', 'D', 1, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 用户管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (100, 1, '用户管理', 'user', 'system/user/index', 'system:user:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 用户管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1001, 100, '用户查询', '', '', 'system:user:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1002, 100, '用户新增', '', '', 'system:user:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1003, 100, '用户修改', '', '', 'system:user:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1004, 100, '用户删除', '', '', 'system:user:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1005, 100, '用户导出', '', '', 'system:user:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1006, 100, '用户导入', '', '', 'system:user:import', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1007, 100, '重置密码', '', '', 'system:user:resetPwd', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 部门管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (101, 1, '部门管理', 'dept', 'system/dept/index', 'system:dept:list', 'ant-design:apartment-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- 部门管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1011, 101, '部门查询', '', '', 'system:dept:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1012, 101, '部门新增', '', '', 'system:dept:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1013, 101, '部门修改', '', '', 'system:dept:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1014, 101, '部门删除', '', '', 'system:dept:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 岗位管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (102, 1, '岗位管理', 'post', 'system/post/index', 'system:post:list', 'ant-design:cluster-outlined', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

-- 岗位管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1021, 102, '岗位查询', '', '', 'system:post:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1022, 102, '岗位新增', '', '', 'system:post:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1023, 102, '岗位修改', '', '', 'system:post:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1024, 102, '岗位删除', '', '', 'system:post:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1025, 102, '岗位导出', '', '', 'system:post:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 菜单管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (103, 1, '菜单管理', 'menu', 'system/menu/index', 'system:menu:list', 'ant-design:menu-outlined', 'M', 4, 1, 1, 0, 0, NOW(), NOW());

-- 菜单管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1031, 103, '菜单查询', '', '', 'system:menu:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1032, 103, '菜单新增', '', '', 'system:menu:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1033, 103, '菜单修改', '', '', 'system:menu:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1034, 103, '菜单删除', '', '', 'system:menu:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 角色管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (104, 1, '角色管理', 'role', 'system/role/index', 'system:role:list', 'ant-design:team-outlined', 'M', 5, 1, 1, 0, 0, NOW(), NOW());

-- 角色管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1041, 104, '角色查询', '', '', 'system:role:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1042, 104, '角色新增', '', '', 'system:role:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1043, 104, '角色修改', '', '', 'system:role:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1044, 104, '角色删除', '', '', 'system:role:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 字典管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (105, 1, '字典管理', 'dict', 'system/dict/index', 'system:dict:list', 'ant-design:book-outlined', 'M', 6, 1, 1, 0, 0, NOW(), NOW());

-- 字典管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1051, 105, '字典查询', '', '', 'system:dict:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1052, 105, '字典新增', '', '', 'system:dict:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1053, 105, '字典修改', '', '', 'system:dict:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1054, 105, '字典删除', '', '', 'system:dict:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1055, 105, '字典导出', '', '', 'system:dict:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 通知公告（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (106, 1, '通知公告', 'notice', 'system/notice/index', 'system:notice:list', 'ant-design:notification-outlined', 'M', 7, 1, 1, 0, 0, NOW(), NOW());

-- 通知公告 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1061, 106, '公告查询', '', '', 'system:notice:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1062, 106, '公告新增', '', '', 'system:notice:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1063, 106, '公告修改', '', '', 'system:notice:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1064, 106, '公告删除', '', '', 'system:notice:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 参数设置（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (107, 1, '参数设置', 'config', 'system/config/index', 'system:config:list', 'ant-design:tool-outlined', 'M', 8, 1, 1, 0, 0, NOW(), NOW());

-- 参数设置 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1071, 107, '参数查询', '', '', 'system:config:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1072, 107, '参数新增', '', '', 'system:config:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1073, 107, '参数修改', '', '', 'system:config:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1074, 107, '参数删除', '', '', 'system:config:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1075, 107, '参数导出', '', '', 'system:config:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 文件管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (108, 1, '文件管理', 'file', 'system/file/index', 'system:file:list', 'ant-design:folder-outlined', 'M', 9, 1, 1, 0, 0, NOW(), NOW());

-- 文件管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1081, 108, '文件查询', '', '', 'system:file:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1082, 108, '文件上传', '', '', 'system:file:upload', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1083, 108, '文件下载', '', '', 'system:file:download', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1084, 108, '文件删除', '', '', 'system:file:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控（目录）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2, 0, '系统监控', 'monitor', '', '', 'ant-design:monitor-outlined', 'D', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 在线用户（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (109, 2, '在线用户', 'online', 'monitor/online/index', 'monitor:online:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 在线用户 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1091, 109, '在线查询', '', '', 'monitor:online:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1092, 109, '强制退出', '', '', 'monitor:online:forceLogout', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 登录日志（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (110, 2, '登录日志', 'loginlog', 'monitor/loginlog/index', 'monitor:loginlog:list', 'ant-design:login-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- 登录日志 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1101, 110, '日志查询', '', '', 'monitor:loginlog:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1102, 110, '日志删除', '', '', 'monitor:loginlog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1103, 110, '日志导出', '', '', 'monitor:loginlog:export', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1104, 110, '清空日志', '', '', 'monitor:loginlog:clear', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 操作日志（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (111, 2, '操作日志', 'operlog', 'monitor/operlog/index', 'monitor:operlog:list', 'ant-design:form-outlined', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

-- 操作日志 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1111, 111, '日志查询', '', '', 'monitor:operlog:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1112, 111, '日志删除', '', '', 'monitor:operlog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1113, 111, '日志导出', '', '', 'monitor:operlog:export', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1114, 111, '清空日志', '', '', 'monitor:operlog:clear', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 服务监控（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (112, 2, '服务监控', 'server', 'monitor/server/index', 'monitor:server:list', 'ant-design:desktop-outlined', 'M', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 缓存监控（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (113, 2, '缓存监控', 'cache', 'monitor/cache/index', 'monitor:cache:list', 'ant-design:database-outlined', 'M', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 缓存列表（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (114, 2, '缓存列表', 'cacheList', 'monitor/cache/list', 'monitor:cache:list', 'ant-design:database-filled', 'M', 6, 1, 1, 0, 0, NOW(), NOW());

-- 系统工具（目录）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3, 0, '系统工具', 'tool', '', '', 'ant-design:tool-outlined', 'D', 3, 1, 1, 0, 0, NOW(), NOW());

-- 系统工具 -> 表单构建（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (115, 3, '表单构建', 'build', 'tool/build/index', '', 'ant-design:build-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 系统工具 -> 代码生成（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (116, 3, '代码生成', 'generator', 'tool/generator/index', 'tool:generator:list', 'ant-design:code-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统工具 -> 系统接口（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (117, 3, '系统接口', 'api', 'tool/api/index', '', 'ant-design:api-outlined', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

-- 关于（目录）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (4, 0, '关于', 'about', '', '', 'ant-design:info-circle-outlined', 'D', 4, 1, 1, 0, 0, NOW(), NOW());

-- 关于 -> API文档（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (118, 4, 'API文档', 'api-docs', 'about/api-docs/index', '', 'ant-design:file-text-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 关于 -> 系统信息（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (119, 4, '系统信息', 'system-info', 'about/system-info/index', '', 'ant-design:desktop-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- ============================================================
-- 关联默认管理员用户与 admin 角色（用户ID=1）
-- ============================================================
INSERT IGNORE INTO sys_user_role (user_id, role_id) VALUES (1, 1);

-- ============================================================
-- 为 admin 角色分配所有菜单权限
-- ============================================================
INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu;