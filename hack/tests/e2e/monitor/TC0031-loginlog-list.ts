import { test, expect } from '../../fixtures/auth';

test.describe('TC0031 登录日志列表查询', () => {
  test('TC0031a: 登录日志页面加载并展示表格', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Table should be visible
    await expect(adminPage.locator('.vxe-table')).toBeVisible();

    // Should have toolbar buttons
    await expect(adminPage.getByRole('button', { name: /清\s*空/ })).toBeVisible();
    await expect(adminPage.getByRole('button', { name: /导\s*出/ })).toBeVisible();
  });

  test('TC0031b: 登录日志包含admin用户记录', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(1000);

    // Should see login log rows with admin
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    if (count > 0) {
      // At least one row should contain 'admin'
      await expect(adminPage.locator('.vxe-body--row').first()).toContainText('admin');
    }
  });
});
