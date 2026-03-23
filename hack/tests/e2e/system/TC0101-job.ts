import { expect, test } from '../../fixtures/auth';

test.describe('定时任务管理', () => {
  test('TC0101: 查询任务列表', async ({ adminPage }) => {
    await adminPage.goto('/system/job');
    await adminPage.waitForLoadState('networkidle');
    await expect(adminPage.locator('text=/会话清理/')).toBeVisible();
    await expect(adminPage.locator('text=/服务器监控/')).toBeVisible();
  });

  test('TC0102: 创建自定义任务', async ({ adminPage }) => {
    await adminPage.goto('/system/job');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.click('button:has-text("新增")');
    await adminPage.fill('input[placeholder*="任务名称"]', '测试任务');
    await adminPage.fill('input[placeholder*="任务分组"]', 'test');
    await adminPage.fill('input[placeholder*="执行指令"]', 'echo test');
    await adminPage.fill('input[placeholder*="Cron"]', '0 0 * * * *');
    await adminPage.click('button:has-text(/确\s*认/)');
    await expect(adminPage.locator('text=/创建成功/')).toBeVisible();
  });

  test('TC0103: 系统任务不可删除', async ({ adminPage }) => {
    await adminPage.goto('/system/job');
    await adminPage.waitForLoadState('networkidle');
    const row = adminPage.locator('tr:has-text("会话清理")');
    await expect(row.locator('button:has-text(/删\s*除/)')).not.toBeVisible();
  });
});
