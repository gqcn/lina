import { test, expect } from '../../fixtures/auth';

test.describe('TC0045 版本信息页面', () => {
  test('TC0045a: 版本信息页面显示四个区块', async ({ adminPage }) => {
    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');

    // 关于项目区块
    await expect(content.getByText('关于项目')).toBeVisible();
    await expect(content.getByText('项目名称')).toBeVisible();
    await expect(content.getByText('版本号')).toBeVisible();
    await expect(content.getByText('v0.5.0')).toBeVisible();

    // 基本信息区块
    await expect(content.getByText('基本信息')).toBeVisible();

    // 后端组件区块
    await expect(content.getByText('后端组件')).toBeVisible();
    await expect(content.getByText('GoFrame', { exact: true })).toBeVisible();

    // 前端组件区块
    await expect(content.getByText('前端组件')).toBeVisible();
    await expect(content.getByText('Vue', { exact: true }).first()).toBeVisible();
  });

  test('TC0045b: 基本信息区块加载后端运行时数据', async ({ adminPage }) => {
    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');

    // Wait for API data to load
    await expect(content.getByText('Go 版本')).toBeVisible({
      timeout: 10_000,
    });
    await expect(content.getByText(/go\d+\.\d+/)).toBeVisible();
    await expect(content.getByText('操作系统')).toBeVisible();
    await expect(content.getByText('启动时间')).toBeVisible();
    await expect(content.getByText('运行时长')).toBeVisible();
  });
});
