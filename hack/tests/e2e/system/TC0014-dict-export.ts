import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0014 字典导出', () => {
  test('TC0014a: 导出字典类型返回正确格式', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/dict/type/export'),
      { timeout: 15000 },
    );

    await dictPage.clickTypeExport();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });

  test('TC0014b: 导出字典数据返回正确格式', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Must select a type first to have data to export
    await dictPage.clickTypeRow('sys_normal_disable');

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/dict/data/export'),
      { timeout: 15000 },
    );

    await dictPage.clickDataExport();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});
