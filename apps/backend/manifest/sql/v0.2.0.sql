-- v0.2.0: Dict Management, Dept Management, Post Management, User-Dept-Post Association

-- ============================================================
-- 字典类型表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_type (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        VARCHAR(128) NOT NULL DEFAULT '',
    type        VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(type)
);

-- ============================================================
-- 字典数据表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_data (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    dict_type   VARCHAR(128) NOT NULL DEFAULT '',
    label       VARCHAR(128) NOT NULL DEFAULT '',
    value       VARCHAR(128) NOT NULL DEFAULT '',
    sort        INTEGER      NOT NULL DEFAULT 0,
    tag_style   VARCHAR(128) NOT NULL DEFAULT '',
    css_class   VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
);

-- ============================================================
-- 部门表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dept (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id   INTEGER      NOT NULL DEFAULT 0,
    ancestors   VARCHAR(512) NOT NULL DEFAULT '',
    name        VARCHAR(128) NOT NULL DEFAULT '',
    order_num   INTEGER      NOT NULL DEFAULT 0,
    leader      INTEGER      NOT NULL DEFAULT 0,
    phone       VARCHAR(20)  NOT NULL DEFAULT '',
    email       VARCHAR(128) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME
);

-- ============================================================
-- 岗位表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_post (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    dept_id     INTEGER      NOT NULL DEFAULT 0,
    code        VARCHAR(128) NOT NULL DEFAULT '',
    name        VARCHAR(128) NOT NULL DEFAULT '',
    sort        INTEGER      NOT NULL DEFAULT 0,
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(code)
);

-- ============================================================
-- 用户-部门关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_dept (
    user_id     INTEGER NOT NULL,
    dept_id     INTEGER NOT NULL,
    PRIMARY KEY (user_id, dept_id)
);

-- ============================================================
-- 用户-岗位关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_post (
    user_id     INTEGER NOT NULL,
    post_id     INTEGER NOT NULL,
    PRIMARY KEY (user_id, post_id)
);

-- ============================================================
-- 字典初始化数据
-- ============================================================

-- 字典类型: 系统开关
INSERT OR IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('系统开关', 'sys_normal_disable', 1, '系统开关列表', datetime('now'), datetime('now'));

-- 字典类型: 用户性别
INSERT OR IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('用户性别', 'sys_user_sex', 1, '用户性别列表', datetime('now'), datetime('now'));

-- 字典数据: 系统开关
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '正常', '1', 1, 'primary', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '停用', '0', 2, 'danger', 1, datetime('now'), datetime('now'));

-- 字典数据: 用户性别
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '男', '1', 1, 'primary', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '女', '2', 2, 'danger', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '未知', '0', 3, 'default', 1, datetime('now'), datetime('now'));

-- ============================================================
-- 部门初始化数据
-- ============================================================
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (1, 0, '0', 'Lina科技', 0, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (2, 1, '0,1', '研发部门', 1, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (3, 1, '0,1', '市场部门', 2, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (4, 1, '0,1', '测试部门', 3, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (5, 1, '0,1', '财务部门', 4, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (6, 1, '0,1', '运维部门', 5, 1, datetime('now'), datetime('now'));

-- ============================================================
-- 岗位初始化数据
-- ============================================================
INSERT OR IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (1, 1, 'CEO', '总经理', 1, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (2, 2, 'CTO', '技术总监', 2, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (3, 2, 'PM', '项目经理', 3, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (4, 2, 'DEV', '开发工程师', 4, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (5, 4, 'QA', '测试工程师', 5, 1, datetime('now'), datetime('now'));

-- ============================================================
-- 关联 admin 用户到 Lina科技 部门和总经理岗位
-- ============================================================
INSERT OR IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (1, 1);
INSERT OR IGNORE INTO sys_user_post (user_id, post_id) VALUES (1, 1);
