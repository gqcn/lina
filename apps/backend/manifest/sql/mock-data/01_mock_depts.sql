-- Mock data: 部门演示数据
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (1, 0, '0', 'Lina科技', 0, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (2, 1, '0,1', '研发部门', 1, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (3, 1, '0,1', '市场部门', 2, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (4, 1, '0,1', '测试部门', 3, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (5, 1, '0,1', '财务部门', 4, 1, datetime('now'), datetime('now'));
INSERT OR IGNORE INTO sys_dept (id, parent_id, ancestors, name, order_num, status, created_at, updated_at)
VALUES (6, 1, '0,1', '运维部门', 5, 1, datetime('now'), datetime('now'));
