import type { Page } from '@playwright/test';

export class UserPage {
  constructor(private page: Page) {}

  /** The Vben drawer (Sheet/Dialog) container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/user');
    await this.page.waitForLoadState('networkidle');
    // Wait for VxeGrid table to render
    await this.page.locator('.vxe-table').waitFor({ state: 'visible', timeout: 10000 });
  }

  async createUser(
    username: string,
    password: string,
    nickname?: string,
  ) {
    // The "新 增" button is in the toolbar (spaced text)
    await this.page.getByRole('button', { name: /新\s*增/ }).click();

    // Wait for drawer (Sheet dialog) to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Fill form fields scoped to the drawer to avoid conflict with the search form
    await this.drawer.getByPlaceholder('请输入用户名').fill(username);
    await this.drawer.getByPlaceholder('请输入密码').fill(password);
    if (nickname) {
      await this.drawer.getByPlaceholder('请输入昵称').fill(nickname);
    }

    // Click the drawer's confirm button (确 认 - note space in Ant Design)
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for API response
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async editUser(username: string, fields: { nickname?: string }) {
    // VXE-Grid with fixed: 'right' action column renders buttons in a separate
    // fixed overlay DOM tree. Search for the user first to narrow to one row.
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    // With search filtering to one row, click the first visible edit button
    // Note: Ant Design adds space between Chinese chars in buttons ("编 辑")
    await this.page.getByRole('button', { name: /编\s*辑/ }).first().click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    if (fields.nickname) {
      const nicknameInput = this.drawer.getByPlaceholder('请输入昵称');
      await nicknameInput.clear();
      await nicknameInput.fill(fields.nickname);
    }

    // Click the drawer's confirm button
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async deleteUser(username: string) {
    // VXE-Grid with fixed: 'right' action column - search to narrow to one row
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    // Click the first visible delete button (ghost-button = ant-btn-sm, not toolbar's full button)
    // Note: Ant Design adds space between Chinese chars in buttons ("删 除")
    await this.page.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Confirm deletion in the Popconfirm
    await this.page.waitForTimeout(500);
    // Popconfirm uses ant-popover
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', { name: /确\s*定|OK|是/i });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      // Fallback: Ant Design Modal confirm
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async hasUser(username: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: username })
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Click a column header to trigger sorting */
  async clickColumnSort(columnTitle: string) {
    // VXE-Grid has duplicate headers (visible + fixed-hidden), use .first() for visible one
    const header = this.page.locator('.vxe-header--column.fixed--visible', { hasText: columnTitle }).first();
    await header.click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Get all cell values for a column by field name */
  async getColumnValues(field: string): Promise<string[]> {
    const cells = this.page.locator(`.vxe-body--column[colid] .vxe-cell`);
    // Use a more reliable way: get all rows and extract the specific column
    const rows = this.page.locator('.vxe-body--row');
    const count = await rows.count();
    const values: string[] = [];
    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      // Try to get the cell text for the column
      const cell = row.locator(`td[field="${field}"] .vxe-cell, td .vxe-cell`);
      // Fallback: use column index mapping
    }
    return values;
  }

  /** Get visible row count */
  async getVisibleRowCount(): Promise<number> {
    return this.page.locator('.vxe-body--row').count();
  }

  /** Fill the search form field by label */
  async fillSearchField(label: string, value: string) {
    // The Vben5 form renders labels as text followed by input fields
    // Use getByLabel which matches aria-label or associated label text
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Select status in search form */
  async selectSearchStatus(statusLabel: string) {
    const form = this.page.locator('.vxe-grid--form-wrapper, .vben-form-wrapper').first();
    const select = form.locator('.ant-select').first();
    await select.click();
    await this.page.getByText(statusLabel, { exact: true }).click();
    await this.page.waitForTimeout(300);
  }

  /** Click search/query button */
  async clickSearch() {
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click reset button */
  async clickReset() {
    await this.page.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click export button */
  async clickExport() {
    await this.page.getByRole('button', { name: /导\s*出/ }).click();
    await this.page.waitForTimeout(2000);
  }

  /** Click import button to open import modal */
  async clickImport() {
    await this.page.getByRole('button', { name: /导\s*入/ }).first().click();
    await this.page.waitForTimeout(500);
  }

  /** Get the total count from the pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator('.vxe-pager--total');
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }
}
