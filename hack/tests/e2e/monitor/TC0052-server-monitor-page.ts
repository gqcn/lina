import { test, expect } from '../../fixtures/auth';

test.describe('TC0052 服务监控页面展示', () => {
  test.beforeEach(async ({ adminPage }) => {
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/server') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/server');
    await responsePromise;
    await adminPage.waitForTimeout(1000);
  });

  test('TC0052a: 服务监控页面加载并展示服务器信息', async ({
    adminPage,
  }) => {
    // Server info card should be visible
    await expect(adminPage.getByText('服务器信息')).toBeVisible();

    // Should show hostname
    await expect(adminPage.getByText('主机名')).toBeVisible();

    // Should show OS info
    await expect(adminPage.getByText('操作系统')).toBeVisible();
  });

  test('TC0052b: CPU指标卡片显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('CPU').first()).toBeVisible();

    // Should show CPU cores count
    await expect(adminPage.getByText(/\d+\s*核/)).toBeVisible();
  });

  test('TC0052c: 内存指标卡片显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('内存').first()).toBeVisible();

    // Should show memory values (e.g., GB)
    await expect(adminPage.getByText(/GB/).first()).toBeVisible();
  });

  test('TC0052d: 磁盘使用表格显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('磁盘使用')).toBeVisible();

    // Should show disk path (like / on unix)
    await expect(adminPage.getByText('/').first()).toBeVisible();
  });

  test('TC0052e: Go运行时信息显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('Go 运行时').first()).toBeVisible();

    // Should show Go version
    await expect(adminPage.getByText(/go\d+\.\d+/)).toBeVisible();

    // Should show Goroutines count
    await expect(adminPage.getByText('Goroutines')).toBeVisible();
  });

  test('TC0052f: 网络流量信息显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('网络流量')).toBeVisible();

    // Should show network labels
    await expect(adminPage.getByText('总发送')).toBeVisible();
    await expect(adminPage.getByText('总接收')).toBeVisible();
  });
});
