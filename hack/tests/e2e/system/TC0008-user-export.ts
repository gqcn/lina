import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0008 用户导出', () => {
  test('TC0008a: 未勾选用户时导出按钮置灰不可点击', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Export button should be visible but disabled
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible();
    await expect(exportBtn).toBeDisabled();
  });

  test('TC0008b: 勾选用户后导出按钮可用且请求包含用户 ID', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Select a non-admin row
    await userPage.selectRow('user001');

    // Export button should now be enabled
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeEnabled();

    const requestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/api/v1/user/export') && req.method() === 'GET',
      { timeout: 15000 },
    );

    await userPage.clickExport();
    const request = await requestPromise;

    expect(request.url()).toContain('/user/export');
    expect(request.url()).toContain('ids');
  });

  test('TC0008c: 导出返回正确的 Content-Type', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.selectRow('user001');

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/user/export'),
      { timeout: 15000 },
    );

    await userPage.clickExport();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
