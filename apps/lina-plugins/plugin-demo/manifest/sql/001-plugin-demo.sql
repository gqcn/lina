-- plugin-demo install sql (MVP)
-- 说明：该文件由宿主插件安装器按需执行；需保证幂等。
-- 约定：一期示例插件无历史兼容负担，菜单与授权种子统一使用 INSERT IGNORE，不做 UPDATE/UPSERT 回填。

DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_ids.id
    FROM (
        SELECT id
        FROM sys_menu
        WHERE menu_key IN ('plugin:plugin-demo:sidebar-entry')
    ) AS menu_ids
);
DELETE FROM sys_menu
WHERE menu_key IN ('plugin:plugin-demo:sidebar-entry');

-- 左侧主菜单顶部示例入口
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (0, 'plugin:plugin-demo:sidebar-entry', '插件示例', 'plugin-demo-sidebar-entry', 'system/plugin/runtime-page', 'plugin-demo:example:view', 'ant-design:appstore-outlined', 'M', -1, 1, 1, 0, 0, '', '插件示例左侧菜单', NOW(), NOW());

INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT 1, id
FROM sys_menu
WHERE menu_key IN (
    'plugin:plugin-demo:sidebar-entry'
);
