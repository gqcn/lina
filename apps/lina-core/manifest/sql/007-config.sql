-- 007: 参数设置模块

-- ----------------------------
-- 1. 参数设置表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `sys_config` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT COMMENT '参数ID',
    `name`       VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT '参数名称',
    `key`        VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT '参数键名',
    `value`      VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT '参数键值',
    `remark`     VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT '备注',
    `created_at` DATETIME         DEFAULT NULL            COMMENT '创建时间',
    `updated_at` DATETIME         DEFAULT NULL            COMMENT '修改时间',
    `deleted_at` DATETIME         DEFAULT NULL            COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='参数设置表';
