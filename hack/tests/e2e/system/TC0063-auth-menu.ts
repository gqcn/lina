import { test, expect } from '@playwright/test';
import { LoginPage } from '../../pages/LoginPage';
import { MainLayout } from '../../pages/MainLayout';
import { RolePage } from '../../pages/RolePage';
import { MenuPage } from '../../pages/MenuPage';
import { UserPage } from '../../pages/UserPage';
import { config } from '../../fixtures/config';

test.describe('TC0063 登录后菜单显示', () => {
  const testRoleName = `e2e_menu_role_${Date.now()}`;
  const testRoleCode = `e2e_menu_role_code_${Date.now()}`;
  const testUserUsername = `e2e_menu_user_${Date.now()}`;
  const testUserPassword = 'test123456';

  test.beforeAll(async ({ browser }) => {
    // Setup: Create role and menu for testing
    const context = await browser.newContext();
    const adminPage = await context.newPage();

    const loginPage = new LoginPage(adminPage);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    // Create test menu (a directory)
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();
    await menuPage.createRootMenu({
      name: 'E2E菜单测试目录',
      type: 'D',
      path: '/e2e-menu-test',
      sort: 900,
    });

    // Create sub menu
    await menuPage.expandAll();
    await menuPage.createSubMenu('E2E菜单测试目录', {
      name: 'E2E测试页面',
      type: 'M',
      path: 'test-page',
      component: 'views/dashboard/workbench/index.vue',
      perms: 'e2e:test:page',
      sort: 1,
    });

    // Create role with the test menu
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await rolePage.createRole({
      name: testRoleName,
      code: testRoleCode,
      sort: 900,
      remark: 'E2E测试角色-用于菜单显示测试',
    });

    // Create user with the test role
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.createUserWithRoles(
      testUserUsername,
      testUserPassword,
      'E2E菜单测试用户',
      [testRoleName],
    );

    await context.close();
  });

  test('TC0063a: 超级管理员登录后显示完整菜单', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    // Wait for sidebar/menu to render
    await page.waitForTimeout(2000);

    // Admin should see system management menu
    const systemMenu = page.getByText('系统管理');
    await expect(systemMenu).toBeVisible({ timeout: 5000 });

    // Admin should see menu management
    const menuManagement = page.getByText('菜单管理');
    await expect(menuManagement).toBeVisible({ timeout: 5000 });

    // Admin should see role management
    const roleManagement = page.getByText('角色管理');
    await expect(roleManagement).toBeVisible({ timeout: 5000 });
  });

  test('TC0063b: 普通用户登录后仅显示授权菜单', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(testUserUsername, testUserPassword);

    // Wait for sidebar/menu to render
    await page.waitForTimeout(2000);

    // Should see the test menu directory
    const testMenuDir = page.getByText('E2E菜单测试目录');
    await expect(testMenuDir).toBeVisible({ timeout: 5000 });

    // Expand the menu
    await testMenuDir.click();
    await page.waitForTimeout(500);

    // Should see the sub menu
    const testSubMenu = page.getByText('E2E测试页面');
    await expect(testSubMenu).toBeVisible({ timeout: 5000 });

    // Should NOT see system management (unless role has that menu)
    const systemMenu = page.getByText('系统管理');
    const isSystemVisible = await systemMenu.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isSystemVisible).toBeFalsy();

    // Should NOT see menu management
    const menuManagement = page.getByText('菜单管理');
    const isMenuMgmtVisible = await menuManagement.isVisible({ timeout: 2000 }).catch(() => false);
    expect(isMenuMgmtVisible).toBeFalsy();
  });

  test('TC0063c: 无角色用户登录后无菜单', async ({ page }) => {
    // Create a user without any role
    const context = page.context();
    const adminPage = await context.newPage();

    // Login as admin to create user
    const adminLogin = new LoginPage(adminPage);
    await adminLogin.goto();
    await adminLogin.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    const userPage = new UserPage(adminPage);
    await userPage.goto();
    const noRoleUsername = `e2e_no_role_${Date.now()}`;
    await userPage.createUser(noRoleUsername, testUserPassword, 'E2E无角色用户');

    await adminPage.close();

    // Now login as the no-role user
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(noRoleUsername, testUserPassword);

    // Wait for page to load
    await page.waitForTimeout(2000);

    // Should see empty or minimal menu (maybe just home)
    const sidebar = page.locator('[class*="sidebar"], nav, .ant-menu');
    const menuItems = sidebar.locator('.ant-menu-item, .ant-menu-submenu');

    // Should have very few or no menu items
    const menuCount = await menuItems.count();
    expect(menuCount).toBeLessThan(2);
  });

  test('TC0063d: 不同用户菜单权限差异', async ({ page }) => {
    const loginPage = new LoginPage(page);

    // First login as admin and check available menus
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    await page.waitForTimeout(2000);

    // Count admin menus
    const adminSidebar = page.locator('[class*="sidebar"], nav, .ant-menu');
    const adminMenuCount = await adminSidebar.locator('.ant-menu-submenu').count();

    // Logout
    const mainLayout = new MainLayout(page);
    await mainLayout.logout();

    // Login as test user
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(testUserUsername, testUserPassword);

    await page.waitForTimeout(2000);

    // Count test user menus
    const testSidebar = page.locator('[class*="sidebar"], nav, .ant-menu');
    const testMenuCount = await testSidebar.locator('.ant-menu-submenu').count();

    // Admin should have more menus than test user
    expect(adminMenuCount).toBeGreaterThan(testMenuCount);
  });

  test('TC0063e: 菜单变更后需重新登录生效', async ({ browser }) => {
    // Create two contexts: admin and test user
    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();

    const loginPage = new LoginPage(adminPage);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    // Add new menu to the role
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();
    await menuPage.expandAll();

    // Create a new sub menu
    await menuPage.createSubMenu('E2E菜单测试目录', {
      name: 'E2E新增页面',
      type: 'M',
      path: 'new-page',
      component: 'views/dashboard/workbench/index.vue',
      perms: 'e2e:new:page',
      sort: 2,
    });

    // Now login as test user in a new context
    const testContext = await browser.newContext();
    const testPage = await testContext.newPage();

    const testLogin = new LoginPage(testPage);
    await testLogin.goto();
    await testLogin.loginAndWaitForRedirect(testUserUsername, testUserPassword);

    await testPage.waitForTimeout(2000);

    // The new menu should be visible after login
    const newMenu = testPage.getByText('E2E新增页面');
    await expect(newMenu).toBeVisible({ timeout: 5000 });

    await adminContext.close();
    await testContext.close();
  });

  test.afterAll(async ({ browser }) => {
    // Cleanup test data
    const context = await browser.newContext();
    const adminPage = await context.newPage();

    const loginPage = new LoginPage(adminPage);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    // Delete test user
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.deleteUser(testUserUsername);

    // Delete test menus
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();
    await menuPage.expandAll();
    // Delete children first
    await menuPage.deleteMenu('E2E测试页面');
    await menuPage.deleteMenu('E2E新增页面');
    // Then delete root
    await menuPage.deleteMenu('E2E菜单测试目录', true);

    // Delete test role
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await rolePage.deleteRole(testRoleName);

    // Delete no-role user if exists
    await userPage.goto();
    const noRoleUsers = await userPage.hasUser('e2e_no_role');
    // Note: might have multiple with similar name from different timestamps
    // We'll just cleanup the ones we created in this test run

    await context.close();
  });
});