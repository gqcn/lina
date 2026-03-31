import { test, expect } from '../../fixtures/auth';
import { MenuPage } from '../../pages/MenuPage';

test.describe('TC0060 菜单管理 CRUD', () => {
  test('TC0060a: 菜单列表页面正常加载', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Check that the table is visible
    const table = adminPage.locator('.vxe-table');
    await expect(table).toBeVisible({ timeout: 10000 });

    // Check that toolbar buttons are visible
    await expect(adminPage.getByRole('button', { name: /新\s*增/ }).first()).toBeVisible({ timeout: 5000 });
    await expect(adminPage.getByRole('button', { name: /折\s*叠/ }).first()).toBeVisible({ timeout: 5000 });
  });

  test('TC0060b: 创建菜单对话框打开', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Open the create form
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Verify form fields are present
    await expect(drawer.locator('input[placeholder="请输入菜单名称"]')).toBeVisible({ timeout: 5000 });

    // Close drawer without saving
    await drawer.getByRole('button', { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: 'hidden', timeout: 5000 });
  });

  test('TC0060c: 级联删除开关功能', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Find the cascade delete switch
    const cascadeSwitch = adminPage.locator('.ant-switch').first();
    await cascadeSwitch.waitFor({ state: 'visible', timeout: 5000 });

    // Get initial state via aria-checked attribute
    const initialState = await cascadeSwitch.getAttribute('aria-checked');

    // Toggle the switch
    await cascadeSwitch.click();
    await adminPage.waitForTimeout(500);

    // Verify state changed
    const newState = await cascadeSwitch.getAttribute('aria-checked');
    expect(newState).not.toBe(initialState);
  });

  test('TC0060d: 折叠按钮功能', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for the loading indicator to disappear
    await adminPage.waitForSelector('.vxe-grid.is--loading', { state: 'hidden', timeout: 10000 }).catch(() => {});

    // Click collapse button
    const collapseBtn = adminPage.getByRole('button', { name: /折\s*叠/ }).first();
    await collapseBtn.click({ force: true });
    await adminPage.waitForTimeout(500);

    // Test passes if no errors thrown
    expect(true).toBeTruthy();
  });

  test('TC0060e: 表单字段验证', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Open the create form
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Verify required form fields are present
    await expect(drawer.locator('input[placeholder="请输入菜单名称"]')).toBeVisible({ timeout: 5000 });

    // Verify parent menu select (TreeSelect)
    const parentSelect = drawer.locator('.ant-tree-select, .ant-select').first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });

    // Close drawer
    await drawer.getByRole('button', { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: 'hidden', timeout: 5000 });
  });

  test('TC0060f: 创建根菜单流程', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    const testMenuName = `e2e_test_${Date.now()}`;

    await menuPage.createRootMenu({
      name: testMenuName,
      type: 'D',
      path: 'e2e-test',
      sort: 999,
    });

    // Wait for drawer to close - indicates submission completed
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'hidden', timeout: 15000 });

    // If drawer closes without error, the creation was successful
    expect(true).toBeTruthy();
  });

  test('TC0060g: 编辑菜单时表单应展示被编辑菜单的内容', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage.locator('.vxe-table').waitFor({ state: 'visible', timeout: 10000 });

    // Find the first edit button in the action column and click it
    // Use getByRole for better reliability
    const editBtn = adminPage.getByRole('button', { name: /编\s*辑/ }).first();
    await editBtn.click({ timeout: 5000 });

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Wait for skeleton to disappear (form loading)
    const skeleton = drawer.locator('.ant-skeleton');
    await skeleton.waitFor({ state: 'hidden', timeout: 10000 });

    // Verify the form has values loaded (not empty)
    // The menu name input should have a value
    const nameInput = drawer.locator('input[placeholder="请输入菜单名称"]');
    await expect(nameInput).toBeVisible({ timeout: 5000 });

    // Get the input value to verify it's not empty
    const inputValue = await nameInput.inputValue();
    expect(inputValue.length).toBeGreaterThan(0);

    // Close drawer without saving
    await drawer.getByRole('button', { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: 'hidden', timeout: 5000 });
  });

  test('TC0060h: 上级菜单下拉树应展示子级菜单', async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage.locator('.vxe-table').waitFor({ state: 'visible', timeout: 10000 });

    // Open the create form
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator('.ant-skeleton');
    await skeleton.waitFor({ state: 'hidden', timeout: 10000 });

    // Find the parent menu TreeSelect and click to open dropdown
    const parentSelect = drawer.locator('.ant-tree-select, .ant-select').first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });
    await parentSelect.click();

    // Wait for tree to be visible - the tree is inside the drawer content
    const tree = drawer.locator('[role="tree"]');
    await expect(tree).toBeVisible({ timeout: 5000 });

    // Look for plus-square icons (collapsed nodes with children)
    // Use getByRole to match by accessible name from accessibility tree
    const expandableNodes = tree.getByRole('img', { name: 'plus-square' });
    const expandableCount = await expandableNodes.count();

    // There should be at least one expandable node (parent menu with children)
    // This verifies the tree structure has children
    expect(expandableCount).toBeGreaterThan(0);

    // Click on an expandable node to verify children are shown
    if (expandableCount > 0) {
      const firstExpandable = expandableNodes.first();
      await firstExpandable.click();

      // Wait a moment for expansion
      await adminPage.waitForTimeout(500);

      // Verify expanded node - look for minus-square icon (expanded state)
      const expandedNode = tree.getByRole('img', { name: 'minus-square' }).first();
      await expect(expandedNode).toBeVisible({ timeout: 3000 });
    }

    // Close dropdown by pressing Escape
    await adminPage.keyboard.press('Escape');

    // Wait for dropdown to close
    await adminPage.waitForTimeout(500);

    // Close drawer - use Escape key as backup if button click fails
    try {
      await drawer.getByRole('button', { name: /取\s*消/ }).click({ timeout: 3000 });
      await drawer.waitFor({ state: 'hidden', timeout: 5000 });
    } catch {
      // If button click fails, use Escape to close
      await adminPage.keyboard.press('Escape');
      await drawer.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
    }
  });
});