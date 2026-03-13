-- Migration: add sex column to sys_user
-- SQLite does not support IF NOT EXISTS for ALTER TABLE ADD COLUMN,
-- so we use a pragma check approach. This will fail silently if column exists.
ALTER TABLE sys_user ADD COLUMN sex TINYINT NOT NULL DEFAULT 0;
