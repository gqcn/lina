-- plugin-demo install sql (MVP)
-- 说明：该文件由宿主插件安装器按需执行；需保证幂等。
-- 约定：一期示例插件无历史兼容负担，菜单与授权种子统一使用 INSERT IGNORE，不做 UPDATE/UPSERT 回填。

CREATE TABLE IF NOT EXISTS plugin_demo_login_audit (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    trace_id    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '链路追踪ID',
    user_name   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
    status      TINYINT      NOT NULL DEFAULT 0 COMMENT '登录状态（0=成功 1=失败）',
    ip          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录IP地址',
    client_type VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '客户端类型（web/mobile/api）',
    message     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审计消息',
    login_time  DATETIME                          COMMENT '登录时间',
    created_at  DATETIME                          COMMENT '创建时间',
    updated_at  DATETIME                          COMMENT '更新时间',
    deleted_at  DATETIME                          COMMENT '删除时间',
    INDEX idx_user_name (user_name),
    INDEX idx_login_time (login_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='示例插件登录审计表';

DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_ids.id
    FROM (
        SELECT id
        FROM sys_menu
        WHERE menu_key IN (
            'plugin:plugin-demo:header-entry',
            'plugin:plugin-demo:login-audit'
        )
    ) AS menu_ids
);
DELETE FROM sys_menu
WHERE menu_key IN (
    'plugin:plugin-demo:header-entry',
    'plugin:plugin-demo:login-audit'
);

-- 左侧主菜单顶部示例入口
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (0, 'plugin:plugin-demo:sidebar-entry', '插件示例', 'plugin-demo-sidebar-entry', 'system/plugin/runtime-page', 'plugin-demo:example:view', 'ant-design:appstore-outlined', 'M', -1, 1, 1, 0, 0, '', '插件示例左侧菜单', NOW(), NOW());

INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT 1, id
FROM sys_menu
WHERE menu_key IN (
    'plugin:plugin-demo:sidebar-entry'
);
