-- v0.2.0: Dict Management, Dept Management, Post Management, User-Dept-Post Association

-- ============================================================
-- 字典类型表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_type (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(128) NOT NULL DEFAULT '',
    type        VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 字典数据表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_data (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    dict_type   VARCHAR(128) NOT NULL DEFAULT '',
    label       VARCHAR(128) NOT NULL DEFAULT '',
    value       VARCHAR(128) NOT NULL DEFAULT '',
    sort        INT          NOT NULL DEFAULT 0,
    tag_style   VARCHAR(128) NOT NULL DEFAULT '',
    css_class   VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 部门表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dept (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    parent_id   INT          NOT NULL DEFAULT 0,
    ancestors   VARCHAR(512) NOT NULL DEFAULT '',
    name        VARCHAR(128) NOT NULL DEFAULT '',
    code        VARCHAR(64)  NOT NULL DEFAULT '',
    order_num   INT          NOT NULL DEFAULT 0,
    leader      INT          NOT NULL DEFAULT 0,
    phone       VARCHAR(20)  NOT NULL DEFAULT '',
    email       VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 岗位表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_post (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    dept_id     INT          NOT NULL DEFAULT 0,
    code        VARCHAR(128) NOT NULL DEFAULT '',
    name        VARCHAR(128) NOT NULL DEFAULT '',
    sort        INT          NOT NULL DEFAULT 0,
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 用户-部门关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_dept (
    user_id     INT NOT NULL,
    dept_id     INT NOT NULL,
    PRIMARY KEY (user_id, dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 用户-岗位关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_post (
    user_id     INT NOT NULL,
    post_id     INT NOT NULL,
    PRIMARY KEY (user_id, post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 字典初始化数据（系统必需）
-- ============================================================

-- 字典类型: 系统开关
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('系统开关', 'sys_normal_disable', 1, '系统开关列表', NOW(), NOW());

-- 字典类型: 用户性别
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('用户性别', 'sys_user_sex', 1, '用户性别列表', NOW(), NOW());

-- 字典数据: 系统开关
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '正常', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '停用', '0', 2, 'danger', 1, NOW(), NOW());

-- 字典数据: 用户性别
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '男', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '女', '2', 2, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '未知', '0', 3, 'default', 1, NOW(), NOW());
