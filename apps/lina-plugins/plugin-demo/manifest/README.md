# plugin-demo manifest

该目录存放插件可发布资源。

## 资源约定

- `../plugin.yaml`：插件统一入口清单，索引本目录下的 SQL 等资源文件
- `sql/001-plugin-demo.sql`：插件安装时执行，命名遵循宿主 `{序号}-{当前迭代名称}.sql` 规范
- `sql/uninstall/001-plugin-demo.sql`：插件卸载时执行（谨慎，建议仅清理插件私有对象）

## 注意

- 一期为源码插件 MVP，宿主初始化流程只扫描 `sql/` 根目录，不会顺序执行 `sql/uninstall/`。
- 插件菜单当前以安装/卸载 SQL 作为单一真相源，并通过 `sys_menu.menu_key` 进行稳定治理，`remark` 只保留备注含义。
- 若宿主已有对应对象，建议使用幂等 SQL（`CREATE TABLE IF NOT EXISTS` / `INSERT IGNORE`）。
