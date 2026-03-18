import { test, expect } from '../../fixtures/auth';

test.describe('TC0046 用户头像下拉菜单', () => {
  test('TC0046a: 下拉菜单不显示文档、Github、问题&帮助', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    // Click the user avatar/name in the header to open the dropdown
    const header = adminPage.locator('header');
    const avatarTrigger = header.locator('button').last();
    await avatarTrigger.click();
    await adminPage.waitForTimeout(500);

    // Get all dropdown menu items text
    const menuItems = adminPage.locator('[role="menuitem"]');
    const count = await menuItems.count();
    const menuTexts: string[] = [];
    for (let i = 0; i < count; i++) {
      const text = await menuItems.nth(i).textContent();
      if (text) menuTexts.push(text.trim());
    }

    // Verify removed menu items do NOT exist
    expect(menuTexts.join(',')).not.toContain('文档');
    expect(menuTexts.join(',')).not.toContain('GitHub');
    expect(menuTexts.join(',')).not.toMatch(/问题/);

    // Verify "个人中心" still exists (with possible ant-design spacing)
    expect(menuTexts.join(',')).toMatch(/个\s*人\s*中\s*心/);
  });

  test('TC0046b: 下拉菜单显示正确的用户昵称和邮箱', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    // Click the user avatar/name in the header to open the dropdown
    const header = adminPage.locator('header');
    const avatarTrigger = header.locator('button').last();
    await avatarTrigger.click();
    await adminPage.waitForTimeout(500);

    // Check that hardcoded "ann.vben@gmail.com" is NOT displayed anywhere
    await expect(
      adminPage.getByText('ann.vben@gmail.com'),
    ).toHaveCount(0);

    // admin user has nickname "管理员", so it should be displayed
    // Use the dropdown content (data-reka-menu-content) to scope
    const dropdownContent = adminPage.locator('[data-reka-menu-content]');
    await expect(dropdownContent).toBeVisible();
    await expect(dropdownContent.getByText(/管\s*理\s*员/)).toBeVisible();
  });

  test('TC0046c: 页面右上角应展示用户头像或默认头像', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    // The header should contain an avatar image (either user's or default)
    const header = adminPage.locator('header');
    const avatarImg = header.locator('img[alt]').first();
    await expect(avatarImg).toBeVisible();

    // The avatar should have a valid src (not empty)
    const src = await avatarImg.getAttribute('src');
    expect(src).toBeTruthy();
    expect(src!.length).toBeGreaterThan(0);
  });
});
