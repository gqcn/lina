import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0056 字典管理导出', () => {
  test('TC0056a: 导出全部字典类型和数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Click export button in type panel
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ }).first();
    await expect(exportBtn).toBeVisible({ timeout: 10000 });
    await exportBtn.click();

    // Verify modal appears with combined export message
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/字典类型.*字典数据/)).toBeVisible();

    // Set up response listener for combined export endpoint
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('dict/export'),
      { timeout: 15000 }
    );

    // Click confirm button
    const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.click();

    // Wait for response and verify
    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });
});
