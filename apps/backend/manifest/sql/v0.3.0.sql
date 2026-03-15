-- v0.3.0: Add dept code field (upgrade from v0.2.0)

-- ============================================================
-- 部门表: 增加部门编码字段（从 v0.2.0 升级时执行）
-- ============================================================
-- Note: For fresh installs, the code column is already in v0.2.0.sql CREATE TABLE.
-- This ALTER is only needed when upgrading from v0.2.0.
-- SQLite does not support IF NOT EXISTS for ALTER TABLE ADD COLUMN,
-- so this will produce a warning on fresh installs (safe to ignore).
ALTER TABLE sys_dept ADD COLUMN code VARCHAR(64) NOT NULL DEFAULT '';

-- 为部门编码创建唯一索引（排除空字符串）
CREATE UNIQUE INDEX IF NOT EXISTS idx_sys_dept_code ON sys_dept(code) WHERE code != '';
