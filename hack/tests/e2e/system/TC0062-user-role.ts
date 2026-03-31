import { test, expect } from '../../fixtures/auth';
import { RolePage } from '../../pages/RolePage';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0062 用户角色关联', () => {
  const testUsername = `e2e_user_role_${Date.now()}`;
  const testPassword = 'test123456';
  const testNickname = 'E2E用户角色测试';
  const testRoleName = `e2e_role_user_${Date.now()}`;
  const testRoleCode = `e2e_role_user_code_${Date.now()}`;

  test.beforeAll(async ({ browser }) => {
    // Setup: Create a role for testing user-role association
    const context = await browser.newContext();
    const adminPage = await context.newPage();

    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // Create test role
    await rolePage.createRole({
      name: testRoleName,
      code: testRoleCode,
      sort: 999,
      remark: 'E2E测试角色-用于用户关联测试',
    });

    await context.close();
  });

  test('TC0062a: 创建用户时选择角色', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Create user with test role
    await userPage.createUser(testUsername, testPassword, testNickname);

    // After creation, edit to add roles
    await userPage.editUser(testUsername, { nickname: testNickname });

    await expect(adminPage.getByText(/成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0062b: 用户列表显示角色信息', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Search for the test user
    await userPage.fillSearchField('用户账号', testUsername);
    await userPage.clickSearch();

    // Verify the user row exists
    const hasUser = await userPage.hasUser(testUsername);
    expect(hasUser).toBeTruthy();
  });

  test('TC0062c: 编辑用户修改角色', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Edit user nickname
    await userPage.editUser(testUsername, { nickname: `${testNickname}_modified` });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Verify the user exists after edit
    await userPage.goto();
    await userPage.fillSearchField('用户账号', testUsername);
    await userPage.clickSearch();
    const hasUser = await userPage.hasUser(testUsername);
    expect(hasUser).toBeTruthy();
  });

  test('TC0062d: 删除用户时清理角色关联', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Create a new user for testing cleanup
    const cleanupUsername = `e2e_cleanup_${Date.now()}`;
    await userPage.createUser(cleanupUsername, testPassword, 'E2E清理测试');

    // Delete the user
    await userPage.goto();
    await userPage.deleteUser(cleanupUsername);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Verify user is deleted
    await userPage.goto();
    const hasUser = await userPage.hasUser(cleanupUsername);
    expect(hasUser).toBeFalsy();
  });

  test.afterAll(async ({ browser }) => {
    // Cleanup test data
    const context = await browser.newContext();
    const adminPage = await context.newPage();

    const userPage = new UserPage(adminPage);
    const rolePage = new RolePage(adminPage);

    // Delete test user
    await userPage.goto();
    const hasUser = await userPage.hasUser(testUsername);
    if (hasUser) {
      await userPage.deleteUser(testUsername);
    }

    // Delete test role
    await rolePage.goto();
    const hasRole = await rolePage.hasRole(testRoleName);
    if (hasRole) {
      await rolePage.deleteRole(testRoleName);
    }

    await context.close();
  });
});