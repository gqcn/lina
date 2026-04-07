-- plugin-demo uninstall sql (MVP)
-- 说明：默认仅删除插件菜单绑定；业务数据是否保留由宿主策略决定。

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
