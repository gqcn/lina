import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0013 字典数据管理 CRUD', () => {
  const testLabel = `测试标签_${Date.now()}`;
  const testValue = `test_val_${Date.now()}`;

  test('TC0013a: 选择字典类型后右侧显示数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Click the sys_normal_disable type row to load data in the right panel
    await dictPage.clickTypeRow('sys_normal_disable');

    // Verify data rows are loaded in the right panel
    const rowCount = await dictPage.getDataRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC0013b: 创建新字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // First select a type to enable the add button
    await dictPage.clickTypeRow('sys_normal_disable');

    await dictPage.createData(testLabel, testValue, { sort: 99 });

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0013c: 编辑字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select the same type
    await dictPage.clickTypeRow('sys_normal_disable');

    await dictPage.editData(testLabel, { label: `${testLabel}修改` });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0013d: 删除字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select the same type
    await dictPage.clickTypeRow('sys_normal_disable');

    await dictPage.deleteData(`${testLabel}修改`);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
