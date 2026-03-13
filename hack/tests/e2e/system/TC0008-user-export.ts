import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0008 用户导出', () => {
  test('TC0008a: 导出请求发送到正确的 API 端点', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const requestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/api/user/export') && req.method() === 'GET',
      { timeout: 15000 },
    );

    await userPage.clickExport();
    const request = await requestPromise;

    expect(request.url()).toContain('/user/export');
  });

  test('TC0008b: 导出返回正确的 Content-Type', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/user/export'),
      { timeout: 15000 },
    );

    await userPage.clickExport();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
