DELETE FROM sys_role_menu
WHERE menu_id IN (
    SELECT menu_ids.id
    FROM (
        SELECT id
        FROM sys_menu
        WHERE menu_key IN ('plugin:plugin-demo-runtime:main-entry')
    ) AS menu_ids
);

DELETE FROM sys_menu
WHERE menu_key IN ('plugin:plugin-demo-runtime:main-entry');

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (
    0,
    'plugin:plugin-demo-runtime:main-entry',
    '运行时插件示例',
    '/plugin-assets/plugin-demo-runtime/v0.1.0/mount.js',
    'system/plugin/runtime-page',
    'plugin-demo-runtime:view',
    'ant-design:deployment-unit-outlined',
    'M',
    -1,
    1,
    1,
    0,
    0,
    '{"pluginAccessMode":"embedded-mount"}',
    'plugin-demo-runtime embedded mount menu entry',
    NOW(),
    NOW()
);

INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT 1, id
FROM sys_menu
WHERE menu_key IN ('plugin:plugin-demo-runtime:main-entry');
