import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0054 字典数据导入', () => {
  test('TC0054a: 选中字典类型后导入按钮可用', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Click on a dict type row to select it
    await dictPage.clickTypeRow('sys_normal_disable');

    // Wait for data to load
    await adminPage.waitForTimeout(500);

    // Import button should be visible in data panel
    const dataPanel = adminPage.locator('#dict-data');
    await expect(dataPanel.getByRole('button', { name: /导\s*入/ })).toBeVisible();
  });

  test('TC0054b: 点击导入按钮打开导入弹窗', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select a dict type first
    await dictPage.clickTypeRow('sys_normal_disable');
    await adminPage.waitForTimeout(500);

    await dictPage.clickDataImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal).toContainText('字典数据导入');
  });

  test('TC0054c: 导入弹窗中有下载模板链接', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeRow('sys_normal_disable');
    await adminPage.waitForTimeout(500);
    await dictPage.clickDataImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText('下载模板')).toBeVisible();
  });

  test('TC0054d: 导入弹窗中有覆盖模式开关', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeRow('sys_normal_disable');
    await adminPage.waitForTimeout(500);
    await dictPage.clickDataImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText(/是否更新\/覆盖已存在的字典数据/)).toBeVisible();
  });

  test('TC0054e: 下载模板请求发送到正确的端点', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeRow('sys_normal_disable');
    await adminPage.waitForTimeout(500);
    await dictPage.clickDataImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/dict/data/import-template'),
      { timeout: 10000 },
    );

    await modal.getByText('下载模板').click();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});