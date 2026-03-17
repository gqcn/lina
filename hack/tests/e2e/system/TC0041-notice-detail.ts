import { test, expect } from '../../fixtures/auth';

test.describe('TC0041 通知公告详情页', () => {
  test('TC0041a: 直接访问通知详情页', async ({ adminPage }) => {
    // Navigate to a mock notice detail page (id=1 from mock data)
    await adminPage.goto('/system/notice/detail/1');
    await adminPage.waitForLoadState('networkidle');

    // Should show the notice title
    await expect(adminPage.getByText('系统升级通知')).toBeVisible({
      timeout: 10000,
    });

    // Should show the notice content
    await expect(
      adminPage.getByText('系统将于本周六凌晨'),
    ).toBeVisible({ timeout: 5000 });
  });

  test('TC0041b: 详情页显示通知类型', async ({ adminPage }) => {
    await adminPage.goto('/system/notice/detail/1');
    await adminPage.waitForLoadState('networkidle');

    // Should show type as dict tag within descriptions
    const descArea = adminPage.locator('.ant-descriptions');
    await expect(descArea).toBeVisible({ timeout: 10000 });
  });

  test('TC0041c: 返回按钮可用', async ({ adminPage }) => {
    await adminPage.goto('/system/notice/detail/1');
    await adminPage.waitForLoadState('networkidle');

    const backButton = adminPage.getByText('返回');
    await expect(backButton).toBeVisible({ timeout: 5000 });
  });
});
