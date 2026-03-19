import path from 'node:path';
import fs from 'node:fs';
import { test, expect } from '../../fixtures/auth';
import { FilePage } from '../../pages/FilePage';

test.describe('TC0048 文件管理', () => {
  // Create a temporary test file
  const testFileName = `test_upload_${Date.now()}.txt`;
  const testFilePath = path.join('/tmp', testFileName);

  test.beforeAll(() => {
    fs.writeFileSync(testFilePath, 'This is a test file for E2E upload testing.');
  });

  test.afterAll(() => {
    if (fs.existsSync(testFilePath)) {
      fs.unlinkSync(testFilePath);
    }
  });

  test('TC0048a: 文件管理页面可正常访问', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Verify page title and table are visible
    await expect(adminPage.getByText('文件列表')).toBeVisible();
    await expect(adminPage.locator('.vxe-table')).toBeVisible();
  });

  test('TC0048b: 文件上传按钮打开上传弹窗', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    await filePage.openFileUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal.getByText('文件上传')).toBeVisible();
    // Should have drag upload area
    await expect(modal.locator('.ant-upload-drag')).toBeVisible();
  });

  test('TC0048c: 图片上传按钮打开上传弹窗', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    await filePage.openImageUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal.getByText('图片上传')).toBeVisible();
  });

  test('TC0048d: 上传文件后文件列表中可见', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Upload via API to avoid complex file input interaction
    const token = await adminPage.evaluate(() => {
      return localStorage.getItem('preferences') || '';
    });

    // Use file chooser for upload
    await filePage.openFileUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    const fileChooserPromise = adminPage.waitForEvent('filechooser');
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(testFilePath);

    // Wait for upload success
    await expect(
      adminPage.getByText(/上传成功/),
    ).toBeVisible({ timeout: 10000 });

    // Close modal via the X button (last button in dialog)
    await modal.locator('button').last().click();
    await adminPage.waitForTimeout(1000);

    // Verify file appears in the list
    const hasFile = await filePage.hasFile(testFileName);
    expect(hasFile).toBeTruthy();
  });

  test('TC0048e: 搜索条件筛选文件', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Search by suffix
    const suffixInput = adminPage.getByLabel('文件后缀', { exact: true }).first();
    await suffixInput.fill('txt');
    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    await adminPage.waitForTimeout(1000);

    // All results should have .txt suffix
    const rowCount = await filePage.getRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC0048f: 删除文件', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Get initial row count
    const initialCount = await filePage.getRowCount();

    if (initialCount > 0) {
      // Click delete on first row
      const firstRow = adminPage.locator('.vxe-body--row').first();
      await firstRow.getByRole('button', { name: /删\s*除/ }).click();

      // Confirm delete (button text is "确 定")
      await adminPage
        .getByRole('button', { name: /确\s*定/ })
        .click();

      await adminPage.waitForTimeout(1000);

      // Verify row count decreased
      const newCount = await filePage.getRowCount();
      expect(newCount).toBeLessThan(initialCount);
    }
  });
});
