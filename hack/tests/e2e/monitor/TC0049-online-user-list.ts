import { test, expect } from '../../fixtures/auth';

test.describe('TC0049 在线用户列表展示', () => {
  test.beforeEach(async ({ adminPage }) => {
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/online/list') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/online');
    await responsePromise;
    await adminPage.waitForTimeout(500);
  });

  test('TC0049a: 在线用户页面加载并展示表格', async ({ adminPage }) => {
    // Table should be visible
    await expect(adminPage.locator('.vxe-table')).toBeVisible();
  });

  test('TC0049b: 工具栏显示在线人数统计', async ({ adminPage }) => {
    // Should show online count text
    await expect(
      adminPage.getByText(/在线用户列表.*共.*人在线/),
    ).toBeVisible();
  });

  test('TC0049c: 表格包含必要的列', async ({ adminPage }) => {
    // Check for expected column headers in the entire header area
    const headerArea = adminPage.locator('.vxe-table--header');
    await expect(headerArea).toContainText('登录账号');
    await expect(headerArea).toContainText('IP地址');
    await expect(headerArea).toContainText('浏览器');
    await expect(headerArea).toContainText('登录时间');
    await expect(headerArea).toContainText('操作');
  });

  test('TC0049d: 当前登录用户出现在列表中', async ({ adminPage }) => {
    // The logged-in admin user should appear in the online list
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);

    // At least one row should contain 'admin'
    await expect(adminPage.locator('.vxe-body--row').first()).toContainText(
      'admin',
    );
  });
});
