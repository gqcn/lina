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
