import type { Page } from '@playwright/test';

export class RolePage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/role');
    await this.page.waitForLoadState('networkidle');
    // Wait for page content to appear - be more flexible
    await this.page.waitForTimeout(2000);
  }

  /** Create a new role by clicking "新增" toolbar button */
  async createRole(params: {
    name: string;
    code: string;
    sort?: number;
    status?: number;
    remark?: string;
  }) {
    // Wait for page to be ready first
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(2000);

    // Click the primary "新增" button in toolbar (use first() as there may be multiple buttons)
    await this.page
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Wait for form to render
    await this.page.waitForTimeout(1500);

    // Fill role name
    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.waitFor({ state: 'visible', timeout: 5000 });
    await nameInput.fill(params.name);

    // Fill role code (permission character)
    const codeInput = this.drawer.locator('input[placeholder="请输入权限字符"]');
    await codeInput.fill(params.code);

    // Fill sort if provided
    if (params.sort !== undefined) {
      const sortInput = this.drawer.locator('input[placeholder="请输入显示顺序"]');
      await sortInput.fill(String(params.sort));
    }

    // Fill remark if provided
    if (params.remark) {
      const remarkInput = this.drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    // Select menus if needed - for basic test we skip menu selection
    // Menu selection is tested separately in TC0061e

    // Click confirm button
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Edit a role: find the row, click edit, update fields in drawer */
  async editRole(roleName: string, newName: string) {
    // Find the row and click the edit button
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /编\s*辑/ }).first().click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Clear and fill the new name
    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Delete a role: find the row, click delete, confirm in Popconfirm */
  async deleteRole(roleName: string) {
    // Find the row and click the delete ghost button
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row
      .locator('.ant-btn-sm')
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Confirm in Popconfirm
    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Check if a role row with the given name is visible */
  async hasRole(roleName: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: roleName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Search role by name */
  async searchRole(name: string) {
    const searchInput = this.page.locator(
      '.vxe-grid--form input[placeholder="请输入角色名称"]',
    );
    await searchInput.fill(name);

    // Click search button
    await this.page
      .locator('.vxe-grid--form')
      .getByRole('button', { name: /搜\s*索/ })
      .click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Reset search */
  async resetSearch() {
    await this.page
      .locator('.vxe-grid--form')
      .getByRole('button', { name: /重\s*置/ })
      .click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Toggle role status */
  async toggleStatus(roleName: string) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    const switchBtn = row.locator('.ant-switch');
    await switchBtn.click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click assign button to go to role-auth page */
  async clickAssign(roleName: string) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /分\s*配/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Check menu in the menu tree table (for role edit) */
  async checkMenu(menuName: string) {
    const menuTree = this.drawer.locator('.vxe-table');
    const menuRow = menuTree.locator('.vxe-body--row', { hasText: menuName });
    const checkbox = menuRow.locator('.vxe-checkbox--icon');
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  /** Uncheck menu in the menu tree table (for role edit) */
  async uncheckMenu(menuName: string) {
    const menuTree = this.drawer.locator('.vxe-table');
    const menuRow = menuTree.locator('.vxe-body--row', { hasText: menuName });
    const checkbox = menuRow.locator('.vxe-checkbox--icon');
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  /** Get checked menu count in drawer */
  async getCheckedMenuCount(): Promise<number> {
    const menuTree = this.drawer.locator('.vxe-table');
    const checkedRows = menuTree.locator('.vxe-body--row.is--checked');
    return await checkedRows.count();
  }

  /** Create role with specific menus */
  async createRoleWithMenus(params: {
    name: string;
    code: string;
    sort?: number;
    remark?: string;
    menuNames?: string[];
  }) {
    await this.page
      .locator('.vxe-grid--toolbar')
      .getByRole('button', { name: /新\s*增/ })
      .click();

    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.fill(params.name);

    const codeInput = this.drawer.locator('input[placeholder="请输入权限字符"]');
    await codeInput.fill(params.code);

    if (params.sort !== undefined) {
      const sortInput = this.drawer.locator('input[placeholder="请输入显示顺序"]');
      await sortInput.fill(String(params.sort));
    }

    if (params.remark) {
      const remarkInput = this.drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    // Select menus if provided
    if (params.menuNames && params.menuNames.length > 0) {
      // Wait for menu tree to render
      await this.drawer.locator('.vxe-table').waitFor({ state: 'visible', timeout: 3000 });
      for (const menuName of params.menuNames) {
        await this.checkMenu(menuName);
      }
    }

    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Assign menus to existing role */
  async assignMenusToRole(roleName: string, menuNames: string[]) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /编\s*辑/ }).first().click();

    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Wait for menu tree
    await this.drawer.locator('.vxe-table').waitFor({ state: 'visible', timeout: 3000 });

    // Clear existing selections - expand all and uncheck all
    const menuTree = this.drawer.locator('.vxe-table');
    const allCheckboxes = menuTree.locator('.vxe-checkbox--icon');
    const count = await allCheckboxes.count();
    for (let i = 0; i < count; i++) {
      const checkbox = allCheckboxes.nth(i);
      const row = checkbox.locator('xpath=..');
      const isChecked = await row.evaluate((el) => el.classList.contains('is--checked'));
      if (isChecked) {
        await checkbox.click();
        await this.page.waitForTimeout(100);
      }
    }

    // Select new menus
    for (const menuName of menuNames) {
      await this.checkMenu(menuName);
    }

    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Navigate to role management page */
  async navigateTo() {
    await this.page.goto('/system/role');
    await this.page.waitForLoadState('networkidle');
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 5000 })
      .catch(() => {});
  }
}