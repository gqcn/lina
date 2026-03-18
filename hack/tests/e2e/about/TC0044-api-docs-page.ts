import { test, expect } from '../../fixtures/auth';

test.describe('TC0044 系统接口页面', () => {
  test('TC0044a: 系统接口页面通过 iframe 加载 Stoplight Elements', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/api-docs');
    // Verify the iframe is visible
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    // Wait for Stoplight Elements to render inside the iframe
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    const apiElement = frame.locator('elements-api');
    await expect(apiElement).toBeAttached({ timeout: 15_000 });
    // Verify Overview is visible in sidebar
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify ENDPOINTS section is visible
    await expect(frame.getByText('ENDPOINTS')).toBeVisible();
  });

  test('TC0044b: 系统接口页面不污染主页面样式', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    // Main page should not have any Stoplight stylesheets injected
    const stoplightLinks = await adminPage
      .locator('link[href*="stoplight/styles"]')
      .count();
    expect(stoplightLinks).toBe(0);
  });

  test('TC0044c: Overview 页面显示 API 标题和描述', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    // Wait for content to load
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify the right panel shows API title and description
    await expect(
      frame.locator('h1', { hasText: 'Lina Admin API' }),
    ).toBeVisible({ timeout: 10_000 });
    await expect(frame.getByText('v0.5.0')).toBeVisible();
  });
});
