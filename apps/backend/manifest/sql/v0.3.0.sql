-- v0.3.0: Operation Log, Login Log, Add dept code field

-- ============================================================
-- 操作日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_oper_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           VARCHAR(50)   NOT NULL DEFAULT '',
    oper_summary    VARCHAR(200)  NOT NULL DEFAULT '',
    oper_type       TINYINT       NOT NULL DEFAULT 0,
    method          VARCHAR(200)  NOT NULL DEFAULT '',
    request_method  VARCHAR(10)   NOT NULL DEFAULT '',
    oper_name       VARCHAR(50)   NOT NULL DEFAULT '',
    oper_url        VARCHAR(500)  NOT NULL DEFAULT '',
    oper_ip         VARCHAR(50)   NOT NULL DEFAULT '',
    oper_param      TEXT          NOT NULL DEFAULT '',
    json_result     TEXT          NOT NULL DEFAULT '',
    status          TINYINT       NOT NULL DEFAULT 0,
    error_msg       TEXT          NOT NULL DEFAULT '',
    cost_time       INTEGER       NOT NULL DEFAULT 0,
    oper_time       DATETIME      NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- 登录日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_login_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_name   VARCHAR(50)  NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 0,
    ip          VARCHAR(50)  NOT NULL DEFAULT '',
    browser     VARCHAR(200) NOT NULL DEFAULT '',
    os          VARCHAR(200) NOT NULL DEFAULT '',
    msg         VARCHAR(500) NOT NULL DEFAULT '',
    login_time  DATETIME     NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- 字典初始化数据：操作类型
-- ============================================================
INSERT OR IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作类型', 'sys_oper_type', 1, '操作日志操作类型列表', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '新增', '1', 1, 'success', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '修改', '2', 2, 'primary', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '删除', '3', 3, 'danger', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导出', '4', 4, 'warning', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导入', '5', 5, 'processing', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '其他', '6', 6, 'default', 1, datetime('now'), datetime('now'));

-- ============================================================
-- 字典初始化数据：操作状态
-- ============================================================
INSERT OR IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作状态', 'sys_oper_status', 1, '操作日志操作状态列表', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '成功', '0', 1, 'success', 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '失败', '1', 2, 'danger', 1, datetime('now'), datetime('now'));

-- ============================================================
-- 部门表: 增加部门编码字段（从 v0.2.0 升级时执行）
-- ============================================================
-- Note: For fresh installs, the code column is already in v0.2.0.sql CREATE TABLE.
-- This ALTER is only needed when upgrading from v0.2.0.
-- SQLite does not support IF NOT EXISTS for ALTER TABLE ADD COLUMN,
-- so this will produce a warning on fresh installs (safe to ignore).
-- Placed at end so the error does not block other statements.
ALTER TABLE sys_dept ADD COLUMN code VARCHAR(64) NOT NULL DEFAULT '';

-- 为部门编码创建唯一索引（排除空字符串）
CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_dept_code ON sys_dept(code) WHERE code != '';
