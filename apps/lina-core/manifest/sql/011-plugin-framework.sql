-- ============================================================
-- 宿主插件表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin (
    id            INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    name          VARCHAR(128) NOT NULL DEFAULT '' COMMENT '插件名称',
    version       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '插件版本号',
    type          VARCHAR(32)  NOT NULL DEFAULT 'source' COMMENT '插件类型（source/wasm/package）',
    runtime       VARCHAR(32)  NOT NULL DEFAULT 'source' COMMENT '运行时类型（source/wasm/package）',
    installed     TINYINT      NOT NULL DEFAULT 0 COMMENT '安装状态（1=已安装 0=未安装）',
    status        TINYINT      NOT NULL DEFAULT 0 COMMENT '启用状态（1=启用 0=禁用）',
    manifest_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '插件清单文件路径',
    checksum      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '插件包校验值',
    installed_at  DATETIME                          COMMENT '安装时间',
    enabled_at    DATETIME                          COMMENT '最后一次启用时间',
    disabled_at   DATETIME                          COMMENT '最后一次禁用时间',
    remark        VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at    DATETIME                          COMMENT '创建时间',
    updated_at    DATETIME                          COMMENT '更新时间',
    deleted_at    DATETIME                          COMMENT '删除时间',
    UNIQUE KEY uk_plugin_id (plugin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件注册表';
