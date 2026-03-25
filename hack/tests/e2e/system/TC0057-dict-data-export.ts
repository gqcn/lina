import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0057 字典数据导出', () => {
  test('TC0057a: 导出全部数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select a dict type row to load dict data in right panel
    await dictPage.clickTypeRow('sys_oper_type');

    // Wait for data to load
    await adminPage.waitForTimeout(500);

    // Click export button in data panel (second one, since there are two export buttons)
    const exportBtns = adminPage.getByRole('button', { name: /导\s*出/ });
    const exportBtn = exportBtns.nth(1); // Data panel is second
    await expect(exportBtn).toBeVisible({ timeout: 10000 });
    await exportBtn.click();

    // Verify modal appears
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/是否导出全部数据/)).toBeVisible();

    // Set up response listener
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('dict/data/export'),
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
