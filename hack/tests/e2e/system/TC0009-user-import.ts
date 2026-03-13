import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0009 用户导入', () => {
  test('TC0009a: 点击导入按钮打开导入弹窗', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.clickImport();

    // Verify modal is visible
    const modal = adminPage.locator('.ant-modal');
    await expect(modal).toBeVisible({ timeout: 3000 });
    await expect(modal.locator('.ant-modal-title')).toContainText('导入用户');
  });

  test('TC0009b: 导入弹窗中有下载模板按钮', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.clickImport();

    const modal = adminPage.locator('.ant-modal');
    await expect(modal.getByRole('button', { name: '下载导入模板' })).toBeVisible();
  });

  test('TC0009c: 导入弹窗中有上传按钮', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.clickImport();

    const modal = adminPage.locator('.ant-modal');
    await expect(modal.locator('button', { hasText: '选择文件并导入' })).toBeVisible();
  });

  test('TC0009d: 下载模板请求发送到正确的端点', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.clickImport();

    const modal = adminPage.locator('.ant-modal');
    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/user/import-template'),
      { timeout: 10000 },
    );

    await modal.getByRole('button', { name: '下载导入模板' }).click();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
