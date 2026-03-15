import { test, expect } from '../../fixtures/auth';

test.describe('TC0029 操作日志导出', () => {
  test('TC0029a: 导出请求返回正确的 Content-Type', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/operlog/export'),
      { timeout: 15000 },
    );

    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await exportBtn.click();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
